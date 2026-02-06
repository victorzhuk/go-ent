package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsZero(t *testing.T) {
	t.Run("int zero", func(t *testing.T) {
		assert.True(t, IsZero(0))
		assert.False(t, IsZero(42))
	})

	t.Run("string zero", func(t *testing.T) {
		assert.True(t, IsZero(""))
		assert.False(t, IsZero("hello"))
	})

	t.Run("pointer zero", func(t *testing.T) {
		var p *int
		assert.True(t, IsZero(p))
		p = new(int)
		assert.False(t, IsZero(p))
	})
}
