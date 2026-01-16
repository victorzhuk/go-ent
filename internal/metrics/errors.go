package metrics

import (
	"errors"
)

var (
	// ErrStoreClosed indicates the store has been closed.
	ErrStoreClosed = errors.New("store closed")

	// ErrInvalidRetention indicates an invalid retention period.
	ErrInvalidRetention = errors.New("invalid retention period")

	// ErrLoadFailed indicates loading from persistence failed.
	ErrLoadFailed = errors.New("load failed")

	// ErrSaveFailed indicates saving to persistence failed.
	ErrSaveFailed = errors.New("save failed")

	// ErrNoSession indicates no session ID in context.
	ErrNoSession = errors.New("no session")

	// ErrSessionNotStarted indicates session was not started.
	ErrSessionNotStarted = errors.New("session not started")

	// ErrInvalidPercentile indicates an invalid percentile value.
	ErrInvalidPercentile = errors.New("invalid percentile")
)
