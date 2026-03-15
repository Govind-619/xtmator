package domain

import "time"

// Project represents a named construction estimation project owned by a User.
type Project struct {
	ID         int64
	UserID     int64
	Name       string
	ClientName string
	Location   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
