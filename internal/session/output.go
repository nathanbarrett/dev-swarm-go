package session

import (
	"sync"
)

// OutputBuffer is a thread-safe ring buffer for output lines
type OutputBuffer struct {
	lines    []OutputLine
	maxLines int
	mu       sync.RWMutex
}

// NewOutputBuffer creates a new output buffer
func NewOutputBuffer(maxLines int) *OutputBuffer {
	return &OutputBuffer{
		lines:    make([]OutputLine, 0, maxLines),
		maxLines: maxLines,
	}
}

// Append adds a line to the buffer
func (b *OutputBuffer) Append(line OutputLine) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.lines) >= b.maxLines {
		// Remove oldest line
		b.lines = b.lines[1:]
	}
	b.lines = append(b.lines, line)
}

// GetAll returns all lines in the buffer
func (b *OutputBuffer) GetAll() []OutputLine {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]OutputLine, len(b.lines))
	copy(result, b.lines)
	return result
}

// GetRecent returns the last n lines
func (b *OutputBuffer) GetRecent(n int) []OutputLine {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if n >= len(b.lines) {
		result := make([]OutputLine, len(b.lines))
		copy(result, b.lines)
		return result
	}

	start := len(b.lines) - n
	result := make([]OutputLine, n)
	copy(result, b.lines[start:])
	return result
}

// Len returns the number of lines in the buffer
func (b *OutputBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.lines)
}

// Clear removes all lines from the buffer
func (b *OutputBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = b.lines[:0]
}
