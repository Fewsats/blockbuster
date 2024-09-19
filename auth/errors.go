package auth

import (
	"errors"
)

var (
	// ErrUserNotFound represents the error when a user is not found.
	ErrUserNotFound = errors.New("user not found")
)
