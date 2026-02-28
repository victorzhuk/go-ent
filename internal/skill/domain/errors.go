package domain

import "fmt"

var (
	// ErrInvalidName is returned when a skill name is empty or invalid.
	ErrInvalidName = fmt.Errorf("invalid name")

	// ErrInvalidDescription is returned when a skill description is empty.
	ErrInvalidDescription = fmt.Errorf("invalid description")

	// ErrNotFound is returned when a skill cannot be found.
	ErrNotFound = fmt.Errorf("not found")

	// ErrDuplicate is returned when attempting to register a duplicate skill.
	ErrDuplicate = fmt.Errorf("duplicate")

	// ErrInvalidTrigger is returned when a trigger is invalid.
	ErrInvalidTrigger = fmt.Errorf("invalid trigger")

	// ErrInvalidVersion is returned when a version is invalid.
	ErrInvalidVersion = fmt.Errorf("invalid version")

	// ErrInvalidCapabilityDesc is returned when a capability description is empty.
	ErrInvalidCapabilityDesc = fmt.Errorf("invalid capability description")

	// ErrInvalidConfidence is returned when a confidence value is outside [0,1].
	ErrInvalidConfidence = fmt.Errorf("confidence must be between 0 and 1")
)
