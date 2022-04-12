package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID   int64
	UUID uuid.UUID
	// Saving the password in clear text for YOUR testing purpose via Postman etc.
	Password          string
	HashedPassword    string
	FullName          string
	Email             string
	PasswordChangedAt *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}
