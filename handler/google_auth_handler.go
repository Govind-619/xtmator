package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/Govind-619/xtmator/usecase"
)

// GoogleAuthHandler handles /api/auth/google and /api/auth/google/callback
type GoogleAuthHandler struct {
	oauth *usecase.GoogleOAuthUsecase

	// In-memory state store (random token → expiry) to prevent CSRF
	mu     sync.Mutex
	states map[string]time.Time
}

func NewGoogleAuthHandler(oauth *usecase.GoogleOAuthUsecase) *GoogleAuthHandler {
	h := &GoogleAuthHandler{
		oauth:  oauth,
		states: make(map[string]time.Time),
	}
	// Purge expired states every 10 minutes
	go func() {
		for range time.Tick(10 * time.Minute) {
			h.purgeExpiredStates()
		}
	}()
	return h
}

// InitiateLogin handles GET /api/auth/google
// Generates a state token and redirects to Google's consent page.
func (h *GoogleAuthHandler) InitiateLogin(w http.ResponseWriter, r *http.Request) {
	state := h.newState()
	url := h.oauth.AuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback handles GET /api/auth/google/callback
// Validates state, exchanges code, issues JWT, redirects to frontend.
func (h *GoogleAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if !h.validateState(state) {
		jsonError(w, "invalid OAuth state — possible CSRF attack", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		jsonError(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	token, _, err := h.oauth.HandleCallback(code)
	if err != nil {
		jsonError(w, "Google login failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to frontend with token in URL fragment (never hits server logs)
	http.Redirect(w, r, "/?token="+token, http.StatusTemporaryRedirect)
}

// ── State management ─────────────────────────────────────────────────────────

func (h *GoogleAuthHandler) newState() string {
	b := make([]byte, 16)
	rand.Read(b)
	state := hex.EncodeToString(b)
	h.mu.Lock()
	h.states[state] = time.Now().Add(10 * time.Minute)
	h.mu.Unlock()
	return state
}

func (h *GoogleAuthHandler) validateState(state string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	exp, ok := h.states[state]
	if !ok {
		return false
	}
	delete(h.states, state)
	return time.Now().Before(exp)
}

func (h *GoogleAuthHandler) purgeExpiredStates() {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	for k, exp := range h.states {
		if now.After(exp) {
			delete(h.states, k)
		}
	}
}
