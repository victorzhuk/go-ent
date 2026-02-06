// Package testutil provides utilities for testing go-ent packages.
//
// This package contains helpers for common testing patterns used throughout
// codebase, particularly for table-driven tests and test setup.
package testutil

import (
	"reflect"
)

// IsZero returns true if v is zero value for its type.
func IsZero[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}
