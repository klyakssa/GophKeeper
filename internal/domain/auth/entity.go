package auth

import "time"

// User struct
type User struct {
	ID        string    `json:"user_id" db:"id"`            // ID is primary key
	Login     string    `json:"login" db:"login"`           // Login is unique
	Password  string    `json:"password" db:"password"`     // Password is hashed
	CreatedAt time.Time `json:"created_at" db:"created_at"` // CreatedAt is timestamp
}
