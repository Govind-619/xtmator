package domain

import "time"

// User represents a registered account.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string // empty for Google-only accounts
	GoogleID     string // Google subject ID, empty for email accounts
	AuthProvider string // "email" | "google"
	CreatedAt    time.Time
}
