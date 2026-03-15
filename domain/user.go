package domain

import "time"

// User represents a registered account.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
