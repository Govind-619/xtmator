package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Govind-619/xtmator/usecase"
)

type contextKey string

const userIDKey contextKey = "userID"

// JWTAuth middleware validates the Authorization: Bearer <token> header.
// On success it injects the userID into the request context.
func JWTAuth(auth *usecase.AuthUsecase, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			jsonError(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		userID, err := auth.ValidateToken(token)
		if err != nil {
			jsonError(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// getUserID extracts the authenticated user ID from context (set by JWTAuth).
func getUserID(r *http.Request) int64 {
	v, _ := r.Context().Value(userIDKey).(int64)
	return v
}
