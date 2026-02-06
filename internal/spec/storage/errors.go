package storage

import "errors"

var (
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrBucketNotFound is returned when a BoltDB bucket doesn't exist.
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrDatabaseClosed is returned when operations are attempted on a closed database.
	ErrDatabaseClosed = errors.New("database closed")
)
