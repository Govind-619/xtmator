package domain

import "time"

// CustomItem represents a user-defined reusable BOQ entry
type CustomItem struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Unit        string    `json:"unit"`
	Rate        float64   `json:"rate"`
	CreatedAt   time.Time `json:"created_at"`
}
