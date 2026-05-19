package models

import "time"

type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"` // Hashed
	CreatedAt time.Time `db:"created_at"`
}
