package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorTypeChecking(t *testing.T) {
	t.Parallel()

	t.Run("ErrInvalidSkillName", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("empty skill name")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidSkillName, err)

		// Test that we can unwrap and check for ErrInvalidSkillName
		assert.True(t, errors.Is(wrapped, ErrInvalidSkillName))
	})

	t.Run("ErrInvalidSkillDescription", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("empty description")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidSkillDescription, err)

		// Test that we can unwrap and check for ErrInvalidSkillDescription
		assert.True(t, errors.Is(wrapped, ErrInvalidSkillDescription))
	})

	t.Run("ErrSkillNotFound", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("skill not found in registry")
		wrapped := fmt.Errorf("%w: %v", ErrSkillNotFound, err)

		// Test that we can unwrap and check for ErrSkillNotFound
		assert.True(t, errors.Is(wrapped, ErrSkillNotFound))
	})

	t.Run("ErrDuplicateSkill", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("skill already registered")
		wrapped := fmt.Errorf("%w: %v", ErrDuplicateSkill, err)

		// Test that we can unwrap and check for ErrDuplicateSkill
		assert.True(t, errors.Is(wrapped, ErrDuplicateSkill))
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

	t.Run("ErrInvalidCapabilityDescription", func(t *testing.T) {
		t.Parallel()

		// Test error wrapping
		err := errors.New("capability description empty")
		wrapped := fmt.Errorf("%w: %v", ErrInvalidCapabilityDescription, err)

		// Test that we can unwrap and check for ErrInvalidCapabilityDescription
		assert.True(t, errors.Is(wrapped, ErrInvalidCapabilityDescription))
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
		assert.Equal(t, "invalid skill name", ErrInvalidSkillName.Error())
		assert.Equal(t, "invalid skill description", ErrInvalidSkillDescription.Error())
		assert.Equal(t, "skill not found", ErrSkillNotFound.Error())
		assert.Equal(t, "duplicate skill", ErrDuplicateSkill.Error())
		assert.Equal(t, "invalid trigger", ErrInvalidTrigger.Error())
		assert.Equal(t, "invalid version", ErrInvalidVersion.Error())
		assert.Equal(t, "invalid capability description", ErrInvalidCapabilityDescription.Error())
		assert.Equal(t, "confidence must be between 0 and 1", ErrInvalidConfidence.Error())

		// Test no trailing punctuation
		assert.False(t, hasTrailingPunctuation(ErrInvalidSkillName.Error()))
		assert.False(t, hasTrailingPunctuation(ErrSkillNotFound.Error()))
		assert.False(t, hasTrailingPunctuation(ErrDuplicateSkill.Error()))
	})
}

func hasTrailingPunctuation(s string) bool {
	return len(s) > 0 && (s[len(s)-1] == '.' || s[len(s)-1] == '!' || s[len(s)-1] == '?')
}
