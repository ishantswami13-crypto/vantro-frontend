package domain

import "time"

// User represents a persisted user record.
type User struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	FullName     *string   `db:"full_name" json:"full_name,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
