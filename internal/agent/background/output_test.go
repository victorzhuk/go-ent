package background

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuffer(t *testing.T) {
	t.Parallel()

	t.Run("creates empty buffer", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		assert.NotNil(t, buf)
		assert.Equal(t, 0, buf.Len())
		assert.Equal(t, "", buf.String())
	})
}

func TestBuffer_Write(t *testing.T) {
	t.Parallel()

	t.Run("writes bytes", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		n, err := buf.Write([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, 5, buf.Len())
		assert.Equal(t, "hello", buf.String())
	})

	t.Run("appends multiple writes", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.Write([]byte("hello"))
		buf.Write([]byte(" "))
		buf.Write([]byte("world"))
		assert.Equal(t, "hello world", buf.String())
	})
}

func TestBuffer_WriteString(t *testing.T) {
	t.Parallel()

	t.Run("writes string", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		n, err := buf.WriteString("hello")
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, "hello", buf.String())
	})

	t.Run("appends multiple strings", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("hello")
		buf.WriteString(" ")
		buf.WriteString("world")
		assert.Equal(t, "hello world", buf.String())
	})
}

func TestBuffer_String(t *testing.T) {
	t.Parallel()

	t.Run("returns empty string for new buffer", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		assert.Equal(t, "", buf.String())
	})

	t.Run("returns buffered content", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("test content")
		assert.Equal(t, "test content", buf.String())
	})
}

func TestBuffer_Read(t *testing.T) {
	t.Parallel()

	t.Run("is alias for String", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("test")
		assert.Equal(t, buf.String(), buf.Read())
	})
}

func TestBuffer_ReadFiltered(t *testing.T) {
	t.Parallel()

	t.Run("returns full content with empty pattern", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("line1\nline2\nline3")
		result, err := buf.ReadFiltered("")
		assert.NoError(t, err)
		assert.Equal(t, "line1\nline2\nline3", result)
	})

	t.Run("filters by regex pattern", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("error: something failed\ninfo: processing\nwarning: low memory")
		result, err := buf.ReadFiltered("error:.*")
		assert.NoError(t, err)
		assert.Equal(t, "error: something failed", result)
	})

	t.Run("filters with multiple matches", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("line 1\nline 2\nline 3")
		result, err := buf.ReadFiltered("line \\d+")
		assert.NoError(t, err)
		assert.Equal(t, "line 1line 2line 3", result)
	})

	t.Run("returns empty when no matches", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("no matches here")
		result, err := buf.ReadFiltered("ERROR.*")
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("returns error for invalid regex", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("test")
		_, err := buf.ReadFiltered("[invalid(")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex")
	})

	t.Run("filters case-insensitive", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("ERROR: failed\nWarning: memory\nerror: retry")
		result, err := buf.ReadFiltered("(?i)error:")
		assert.NoError(t, err)
		assert.Contains(t, result, "ERROR:")
		assert.Contains(t, result, "error:")
	})
}

func TestBuffer_Len(t *testing.T) {
	t.Parallel()

	t.Run("returns zero for new buffer", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		assert.Equal(t, 0, buf.Len())
	})

	t.Run("returns byte count", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("hello")
		assert.Equal(t, 5, buf.Len())
		buf.WriteString(" world")
		assert.Equal(t, 11, buf.Len())
	})
}

func TestBuffer_Reset(t *testing.T) {
	t.Parallel()

	t.Run("clears buffer content", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("some content")
		assert.Equal(t, 12, buf.Len())

		buf.Reset()
		assert.Equal(t, 0, buf.Len())
		assert.Equal(t, "", buf.String())
	})
}

func TestBuffer_ThreadSafety(t *testing.T) {
	t.Parallel()

	t.Run("concurrent writes are safe", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				buf.WriteString(string(rune('a' + n)))
			}(i)
		}

		wg.Wait()
		result := buf.String()
		assert.Equal(t, 10, len(result))
	})

	t.Run("concurrent reads are safe", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("test content")

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = buf.String()
			}()
		}

		wg.Wait()
		assert.Equal(t, "test content", buf.String())
	})

	t.Run("concurrent reads and writes are safe", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		var wg sync.WaitGroup

		for i := 0; i < 5; i++ {
			wg.Add(2)
			go func(n int) {
				defer wg.Done()
				buf.WriteString(string(rune('a' + n)))
			}(i)
			go func() {
				defer wg.Done()
				_ = buf.String()
			}()
		}

		wg.Wait()
		result := buf.String()
		assert.Equal(t, 5, len(result))
	})

	t.Run("concurrent filtered reads are safe", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		buf.WriteString("line1\nline2\nline3\nline4\nline5")

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = buf.ReadFiltered("line[135]")
			}()
		}

		wg.Wait()
	})
}

func TestBuffer_Integration(t *testing.T) {
	t.Parallel()

	t.Run("simulates streaming output", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()

		streams := []string{
			"Starting process...\n",
			"Loading config...\n",
			"Processing data...\n",
			"Complete!\n",
		}

		for _, s := range streams {
			buf.WriteString(s)
		}

		full := buf.String()
		expected := "Starting process...\nLoading config...\nProcessing data...\nComplete!\n"
		assert.Equal(t, expected, full)
	})

	t.Run("filters only errors from mixed output", func(t *testing.T) {
		t.Parallel()

		buf := NewBuffer()
		output := `INFO: Starting task
DEBUG: Config loaded
ERROR: Connection failed
INFO: Retrying...
ERROR: Timeout exceeded
SUCCESS: Task complete
`

		buf.WriteString(output)

		errors, err := buf.ReadFiltered("(?m)^ERROR:.*$")
		assert.NoError(t, err)
		assert.Contains(t, errors, "ERROR: Connection failed")
		assert.Contains(t, errors, "ERROR: Timeout exceeded")
		assert.NotContains(t, errors, "INFO: Starting task")
	})
}
