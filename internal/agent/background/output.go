package background

import (
	"bytes"
	"fmt"
	"regexp"
	"sync"
)

// Buffer is a thread-safe output buffer for streaming agent output.
type Buffer struct {
	mu  sync.RWMutex
	buf *bytes.Buffer
}

// NewBuffer creates a new output buffer.
func NewBuffer() *Buffer {
	return &Buffer{
		buf: &bytes.Buffer{},
	}
}

// Write appends data to the buffer.
func (b *Buffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(data)
}

// WriteString appends a string to the buffer.
func (b *Buffer) WriteString(s string) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.WriteString(s)
}

// String returns the full buffer contents.
func (b *Buffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.buf.String()
}

// Read returns the full buffer contents (alias for String).
func (b *Buffer) Read() string {
	return b.String()
}

// ReadFiltered returns buffer contents filtered by the given regex pattern.
// If pattern is empty or invalid, returns the full buffer contents.
func (b *Buffer) ReadFiltered(pattern string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if pattern == "" {
		return b.buf.String(), nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	var result []byte
	input := b.buf.Bytes()
	matches := re.FindAll(input, -1)

	for _, match := range matches {
		result = append(result, match...)
	}

	return string(result), nil
}

// Len returns the current buffer size in bytes.
func (b *Buffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.buf.Len()
}

// Reset clears the buffer contents.
func (b *Buffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}
