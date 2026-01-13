package ast

import "errors"

var (
	// ErrInvalidSource indicates the Go source code is invalid or malformed.
	ErrInvalidSource = errors.New("invalid source")

	// ErrFileNotFound indicates the source file was not found.
	ErrFileNotFound = errors.New("file not found")

	// ErrParseFailed indicates parsing the AST failed.
	ErrParseFailed = errors.New("parse failed")

	// ErrEmptySource indicates the source code is empty.
	ErrEmptySource = errors.New("empty source")
)
