package parser

import "errors"

var (
	ErrInvalidFormat  = errors.New("invalid format")
	ErrInvalidTaskID  = errors.New("invalid task ID")
	ErrDuplicateID    = errors.New("duplicate task ID")
	ErrInvalidStatus  = errors.New("invalid status")
	ErrInvalidYAML    = errors.New("invalid YAML metadata")
	ErrEmptyContent   = errors.New("empty content")
	ErrMissingTaskNum = errors.New("missing task number")
	ErrInvalidDepends = errors.New("invalid dependency format")
)
