package usecase

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthUsecase handles registration, login, and token operations.
type AuthUsecase struct {
	users  repository.UserRepository
	secret []byte

	// Login lockout tracking
	mu       sync.Mutex
	failures map[string]*lockoutState // key = email
}

type lockoutState struct {
	count    int
	lockedAt time.Time
}

const (
	maxFailedAttempts = 5
	lockoutDuration   = 15 * time.Minute
	minPasswordLength = 8
)

func NewAuthUsecase(users repository.UserRepository) *AuthUsecase {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "xtmator-dev-secret-change-in-prod"
	}
	return &AuthUsecase{
		users:    users,
		secret:   []byte(secret),
		failures: make(map[string]*lockoutState),
	}
}

// Register creates a new user. Validates email format and password strength.
func (a *AuthUsecase) Register(name, email, password string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))

	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email and password are required")
	}
	if !isValidEmail(email) {
		return nil, errors.New("invalid email format")
	}
	if err := checkPasswordStrength(password); err != nil {
		return nil, err
	}
	existing, err := a.users.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return a.users.Create(name, email, string(hash))
}

// Login verifies credentials with lockout protection and returns a JWT.
func (a *AuthUsecase) Login(email, password string) (string, *domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	// Check lockout
	if err := a.checkLockout(email); err != nil {
		return "", nil, err
	}

	user, err := a.users.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user == nil || user.PasswordHash == "" {
		a.recordFailure(email)
		return "", nil, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		a.recordFailure(email)
		return "", nil, errors.New("invalid email or password")
	}

	a.clearFailures(email)
	token, _, err := a.issueToken(user)
	return token, user, err
}

// issueToken creates a signed JWT for the given user (shared with google_oauth.go).
func (a *AuthUsecase) issueToken(user *domain.User) (string, *domain.User, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Name,
		"exp":  time.Now().Add(4 * time.Hour).Unix(),
	})
	signed, err := tok.SignedString(a.secret)
	if err != nil {
		return "", nil, fmt.Errorf("sign token: %w", err)
	}
	return signed, user, nil
}

// ValidateToken parses and validates a JWT, returning the user ID.
func (a *AuthUsecase) ValidateToken(tokenStr string) (int64, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return a.secret, nil
	})
	if err != nil || !tok.Valid {
		return 0, errors.New("invalid or expired token")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid subject claim")
	}
	return int64(sub), nil
}

// ── Lockout helpers ──────────────────────────────────────────────────────────

func (a *AuthUsecase) checkLockout(email string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	st, ok := a.failures[email]
	if !ok {
		return nil
	}
	if st.count >= maxFailedAttempts {
		if time.Since(st.lockedAt) < lockoutDuration {
			remaining := lockoutDuration - time.Since(st.lockedAt)
			return fmt.Errorf("account locked due to too many failed attempts — try again in %d minutes",
				int(remaining.Minutes())+1)
		}
		delete(a.failures, email)
	}
	return nil
}

func (a *AuthUsecase) recordFailure(email string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	st := a.failures[email]
	if st == nil {
		st = &lockoutState{}
		a.failures[email] = st
	}
	st.count++
	if st.count >= maxFailedAttempts {
		st.lockedAt = time.Now()
	}
}

func (a *AuthUsecase) clearFailures(email string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.failures, email)
}

// ── Validation helpers ───────────────────────────────────────────────────────

func isValidEmail(email string) bool {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	local, domain := parts[0], parts[1]
	return len(local) > 0 && strings.Contains(domain, ".") && len(domain) > 2
}

func checkPasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, and one number")
	}
	return nil
}

// GetUser retrieves a user by their ID
func (a *AuthUsecase) GetUser(id int64) (*domain.User, error) {
	return a.users.FindByID(id)
}
