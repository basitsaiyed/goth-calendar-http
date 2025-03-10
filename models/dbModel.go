package models

import (
	"time"
)

// User represents a user record in the database
type User struct {
	ID           int       `json:"id"`
	GoogleID     string    `json:"google_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenExpiry  time.Time `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
