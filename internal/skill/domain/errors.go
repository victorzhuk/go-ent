package domain

import "errors"

var (
	ErrInvalidName           = errors.New("invalid name")
	ErrInvalidDescription    = errors.New("invalid description")
	ErrNotFound              = errors.New("not found")
	ErrDuplicate             = errors.New("duplicate")
	ErrInvalidTrigger        = errors.New("invalid trigger")
	ErrInvalidVersion        = errors.New("invalid version")
	ErrInvalidCapabilityDesc = errors.New("invalid capability description")
	ErrInvalidConfidence     = errors.New("confidence must be between 0 and 1")
)
