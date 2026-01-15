package background

import "errors"

var (
	// ErrAgentNotFound is returned when an agent cannot be found.
	ErrAgentNotFound = errors.New("agent not found")
)
