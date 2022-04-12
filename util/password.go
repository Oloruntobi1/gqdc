package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrPasswordMismatch is returned when the password does not match
	ErrPasswordMismatch = fmt.Errorf("password mismatch")
)

// HashPassword returns the bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword checks if the provided password is correct or not
func CheckPassword(password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return ErrPasswordMismatch
		}
		return fmt.Errorf("failed to check password: %w", err)
	}
	return nil
}
