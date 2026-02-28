package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsZero(t *testing.T) {
	t.Parallel()

	t.Run("int zero", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsZero(0))
		assert.False(t, IsZero(42))
	})

	t.Run("string zero", func(t *testing.T) {
		t.Parallel()
		assert.True(t, IsZero(""))
		assert.False(t, IsZero("hello"))
	})

	t.Run("pointer zero", func(t *testing.T) {
		t.Parallel()
		var p *int
		assert.True(t, IsZero(p))
		p = new(int)
		assert.False(t, IsZero(p))
	})
}
