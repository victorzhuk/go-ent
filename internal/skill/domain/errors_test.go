package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorTypeChecking(t *testing.T) {
	t.Parallel()

	t.Run("ErrInvalidName", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("empty skill name")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidName, err)

		// Test that we can unwrap and check for ErrInvalidName
		assert.True(t, errors.Is(wrapped, ErrInvalidName))
	})

	t.Run("ErrInvalidDescription", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("empty description")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidDescription, err)

		// Test that we can unwrap and check for ErrInvalidDescription
		assert.True(t, errors.Is(wrapped, ErrInvalidDescription))
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("skill not found in registry")
		wrapped := fmt.Errorf("%w: %v", ErrNotFound, err)

		// Test that we can unwrap and check for ErrNotFound
		assert.True(t, errors.Is(wrapped, ErrNotFound))
	})

	t.Run("ErrDuplicate", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("skill already registered")
		wrapped := fmt.Errorf("%w: %v", ErrDuplicate, err)

		// Test that we can unwrap and check for ErrDuplicate
		assert.True(t, errors.Is(wrapped, ErrDuplicate))
	})

	t.Run("ErrInvalidTrigger", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("invalid trigger pattern")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidTrigger, err)

		// Test that we can unwrap and check for ErrInvalidTrigger
		assert.True(t, errors.Is(wrapped, ErrInvalidTrigger))
	})

	t.Run("ErrInvalidVersion", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("version format invalid")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidVersion, err)

		// Test that we can unwrap and check for ErrInvalidVersion
		assert.True(t, errors.Is(wrapped, ErrInvalidVersion))
	})

	t.Run("ErrInvalidCapabilityDesc", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("capability description empty")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidCapabilityDesc, err)

		// Test that we can unwrap and check for ErrInvalidCapabilityDesc
		assert.True(t, errors.Is(wrapped, ErrInvalidCapabilityDesc))
	})

	t.Run("ErrInvalidConfidence", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("confidence out of range")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidConfidence, err)

		// Test that we can unwrap and check for ErrInvalidConfidence
		assert.True(t, errors.Is(wrapped, ErrInvalidConfidence))
	})

	t.Run("wrapped error checking", func(t *testing.T) {
		t.Parallel()

		// Create a wrapped error chain
		baseErr := errors.New("validation failed")
		wrapped := fmt.Errorf("create skill: %w", baseErr)
		fullyWrapped := fmt.Errorf("register: %w", wrapped)

		// Test that we can check for base error
		assert.True(t, errors.Is(fullyWrapped, baseErr))
	})

	t.Run("error messages", func(t *testing.T) {
		t.Parallel()

		// Test that error messages are lowercase
		assert.Equal(t, "invalid name", ErrInvalidName.Error())
		assert.Equal(t, "invalid description", ErrInvalidDescription.Error())
		assert.Equal(t, "not found", ErrNotFound.Error())
		assert.Equal(t, "duplicate", ErrDuplicate.Error())
		assert.Equal(t, "invalid trigger", ErrInvalidTrigger.Error())
		assert.Equal(t, "invalid version", ErrInvalidVersion.Error())
		assert.Equal(t, "invalid capability description", ErrInvalidCapabilityDesc.Error())
		assert.Equal(t, "confidence must be between 0 and 1", ErrInvalidConfidence.Error())

		// Test no trailing punctuation
		assert.False(t, hasTrailingPunctuation(ErrInvalidName.Error()))
		assert.False(t, hasTrailingPunctuation(ErrNotFound.Error()))
		assert.False(t, hasTrailingPunctuation(ErrDuplicate.Error()))
	})
}

func hasTrailingPunctuation(s string) bool {
	return len(s) > 0 && (s[len(s)-1] == '.' || s[len(s)-1] == '!' || s[len(s)-1] == '?')
}
