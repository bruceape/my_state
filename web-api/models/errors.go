package models

import "errors"

var (
	// ErrUserExists indicates a user cannot be created because the unique constraint (email) was violated.
	ErrUserExists = errors.New("user exists")

	// ErrUserNotFound indicates a user lookup by email (or ID) returned no result.
	ErrUserNotFound = errors.New("user not found")
)
