package domain

import "time"

// Project represents a named construction estimation project owned by a User.
type Project struct {
	ID         int64
	UserID     int64
	Name       string
	ClientName string
	Location   string
	CostIndex  float64
	ShareToken string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ProjectSheet struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
