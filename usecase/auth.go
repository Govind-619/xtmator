package usecase

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthUsecase handles registration, login, and token validation.
type AuthUsecase struct {
	users  repository.UserRepository
	secret []byte
}

// NewAuthUsecase creates an AuthUsecase. JWT secret is read from the JWT_SECRET env var,
// falling back to a local default (override in production!).
func NewAuthUsecase(users repository.UserRepository) *AuthUsecase {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "xtmator-dev-secret-change-in-prod"
	}
	return &AuthUsecase{users: users, secret: []byte(secret)}
}

// Register creates a new user account. Returns an error if the email already exists.
func (a *AuthUsecase) Register(name, email, password string) (*domain.User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email and password are required")
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

// Login verifies credentials and returns a signed JWT on success.
func (a *AuthUsecase) Login(email, password string) (string, *domain.User, error) {
	user, err := a.users.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid email or password")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", nil, fmt.Errorf("sign token: %w", err)
	}
	return signed, user, nil
}

// ValidateToken parses and validates a JWT, returning the user ID it encodes.
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
