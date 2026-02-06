package storage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorTypeChecking(t *testing.T) {
	t.Parallel()

	t.Run("ErrNotFound", func(t *testing.T) {
		t.Parallel()

		// Test that ErrNotFound is a sentinel error
		assert.True(t, errors.Is(ErrNotFound, ErrNotFound))

		// Test that we can check for ErrNotFound
		testErr := ErrNotFound
		assert.True(t, errors.Is(testErr, ErrNotFound))
	})

	t.Run("ErrInvalidInput", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("nil task")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidInput, err)

		// Test that we can unwrap and check for ErrInvalidInput
		assert.True(t, errors.Is(wrapped, ErrInvalidInput))
	})

	t.Run("ErrBucketNotFound", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("tasks bucket not found")
		wrapped := fmt.Errorf("%w: %v", ErrBucketNotFound, err)

		// Test that we can unwrap and check for ErrBucketNotFound
		assert.True(t, errors.Is(wrapped, ErrBucketNotFound))
	})

	t.Run("ErrDatabaseClosed", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("database closed")
		wrapped := fmt.Errorf("%w: %v", ErrDatabaseClosed, err)

		// Test that we can unwrap and check for ErrDatabaseClosed
		assert.True(t, errors.Is(wrapped, ErrDatabaseClosed))
	})

	t.Run("wrapped error checking", func(t *testing.T) {
		t.Parallel()

		// Create a wrapped error chain
		baseErr := errors.New("database operation failed")
		wrapped := fmt.Errorf("get task: %w", baseErr)
		fullyWrapped := fmt.Errorf("%w: tasks", wrapped)

		// Test that we can check for the base error
		assert.True(t, errors.Is(fullyWrapped, baseErr))
	})

	t.Run("error messages", func(t *testing.T) {
		t.Parallel()

		// Test that error messages are lowercase
		assert.Equal(t, "not found", ErrNotFound.Error())
		assert.Equal(t, "invalid input", ErrInvalidInput.Error())
		assert.Equal(t, "bucket not found", ErrBucketNotFound.Error())
		assert.Equal(t, "database closed", ErrDatabaseClosed.Error())

		// Test no trailing punctuation
		assert.False(t, hasTrailingPunctuation(ErrNotFound.Error()))
		assert.False(t, hasTrailingPunctuation(ErrInvalidInput.Error()))
		assert.False(t, hasTrailingPunctuation(ErrBucketNotFound.Error()))
		assert.False(t, hasTrailingPunctuation(ErrDatabaseClosed.Error()))
	})
}

func hasTrailingPunctuation(s string) bool {
	return len(s) > 0 && (s[len(s)-1] == '.' || s[len(s)-1] == '!' || s[len(s)-1] == '?')
}
