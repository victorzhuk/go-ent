package domain

import "fmt"

var (
	// ErrInvalidSkillName is returned when a skill name is empty or invalid.
	ErrInvalidSkillName = fmt.Errorf("invalid skill name")

	// ErrInvalidSkillDescription is returned when a skill description is empty.
	ErrInvalidSkillDescription = fmt.Errorf("invalid skill description")

	// ErrSkillNotFound is returned when a skill cannot be found.
	ErrSkillNotFound = fmt.Errorf("skill not found")

	// ErrDuplicateSkill is returned when attempting to register a duplicate skill.
	ErrDuplicateSkill = fmt.Errorf("duplicate skill")

	// ErrInvalidTrigger is returned when a trigger is invalid.
	ErrInvalidTrigger = fmt.Errorf("invalid trigger")

	// ErrInvalidVersion is returned when a version is invalid.
	ErrInvalidVersion = fmt.Errorf("invalid version")

	// ErrInvalidCapabilityDescription is returned when a capability description is empty.
	ErrInvalidCapabilityDescription = fmt.Errorf("invalid capability description")

	// ErrInvalidConfidence is returned when a confidence value is outside [0,1].
	ErrInvalidConfidence = fmt.Errorf("confidence must be between 0 and 1")
)
