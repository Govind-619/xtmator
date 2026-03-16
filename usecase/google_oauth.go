package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Govind-619/xtmator/domain"
	"github.com/Govind-619/xtmator/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuthUsecase handles the Google OAuth2 flow.
type GoogleOAuthUsecase struct {
	config *oauth2.Config
	users  repository.UserRepository
	auth   *AuthUsecase
}

type googleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// NewGoogleOAuthUsecase builds the OAuth usecase from environment variables.
func NewGoogleOAuthUsecase(users repository.UserRepository, auth *AuthUsecase) *GoogleOAuthUsecase {
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:3333/api/auth/google/callback"
	}
	cfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &GoogleOAuthUsecase{config: cfg, users: users, auth: auth}
}

// AuthURL returns the Google consent page URL for the given state token.
func (g *GoogleOAuthUsecase) AuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// HandleCallback exchanges the auth code for a user, creates/finds the account,
// and returns a signed JWT ready for the frontend.
func (g *GoogleOAuthUsecase) HandleCallback(code string) (string, *domain.User, error) {
	tok, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		return "", nil, fmt.Errorf("exchange code: %w", err)
	}

	// Fetch user info from Google
	client := g.config.Client(context.Background(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", nil, fmt.Errorf("fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", nil, fmt.Errorf("decode userinfo: %w", err)
	}
	if info.Email == "" {
		return "", nil, errors.New("no email returned from Google")
	}

	// Find or create user in DB
	user, err := g.users.FindOrCreateGoogleUser(info.Name, info.Email, info.ID)
	if err != nil {
		return "", nil, fmt.Errorf("find or create user: %w", err)
	}

	// Issue JWT via the regular auth usecase
	jwtToken, _, err := g.auth.issueToken(user)
	if err != nil {
		return "", nil, err
	}
	return jwtToken, user, nil
}
