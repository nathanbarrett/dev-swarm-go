package session

import (
	"sync"
	"testing"
	"time"
)

func TestNewOutputBuffer(t *testing.T) {
	buf := NewOutputBuffer(100)

	if buf == nil {
		t.Fatal("NewOutputBuffer returned nil")
	}
	if buf.maxLines != 100 {
		t.Errorf("maxLines = %d, want 100", buf.maxLines)
	}
	if len(buf.lines) != 0 {
		t.Errorf("initial len = %d, want 0", len(buf.lines))
	}
}

func TestOutputBufferAppend(t *testing.T) {
	buf := NewOutputBuffer(5)

	// Add some lines
	for i := 0; i < 3; i++ {
		buf.Append(OutputLine{
			Timestamp: time.Now(),
			Text:      "line",
			Stream:    "stdout",
		})
	}

	if buf.Len() != 3 {
		t.Errorf("Len() = %d, want 3", buf.Len())
	}
}

func TestOutputBufferAppendOverflow(t *testing.T) {
	buf := NewOutputBuffer(3)

	// Add more lines than capacity
	for i := 0; i < 5; i++ {
		buf.Append(OutputLine{
			Timestamp: time.Now(),
			Text:      string(rune('A' + i)),
			Stream:    "stdout",
		})
	}

	// Should only have 3 lines
	if buf.Len() != 3 {
		t.Errorf("Len() = %d, want 3", buf.Len())
	}

	// Should have the last 3 lines (C, D, E)
	lines := buf.GetAll()
	if lines[0].Text != "C" {
		t.Errorf("lines[0].Text = %q, want %q", lines[0].Text, "C")
	}
	if lines[1].Text != "D" {
		t.Errorf("lines[1].Text = %q, want %q", lines[1].Text, "D")
	}
	if lines[2].Text != "E" {
		t.Errorf("lines[2].Text = %q, want %q", lines[2].Text, "E")
	}
}

func TestOutputBufferGetAll(t *testing.T) {
	buf := NewOutputBuffer(10)

	// Add lines
	for i := 0; i < 5; i++ {
		buf.Append(OutputLine{
			Text:   string(rune('A' + i)),
			Stream: "stdout",
		})
	}

	lines := buf.GetAll()

	if len(lines) != 5 {
		t.Errorf("len(GetAll()) = %d, want 5", len(lines))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect buffer)
	lines[0].Text = "modified"
	originalLines := buf.GetAll()
	if originalLines[0].Text == "modified" {
		t.Error("GetAll() should return a copy")
	}
}

func TestOutputBufferGetRecent(t *testing.T) {
	buf := NewOutputBuffer(10)

	// Add 5 lines
	for i := 0; i < 5; i++ {
		buf.Append(OutputLine{
			Text:   string(rune('A' + i)),
			Stream: "stdout",
		})
	}

	// Get last 3
	recent := buf.GetRecent(3)
	if len(recent) != 3 {
		t.Errorf("len(GetRecent(3)) = %d, want 3", len(recent))
	}
	if recent[0].Text != "C" {
		t.Errorf("recent[0].Text = %q, want %q", recent[0].Text, "C")
	}
	if recent[2].Text != "E" {
		t.Errorf("recent[2].Text = %q, want %q", recent[2].Text, "E")
	}

	// Get more than available
	all := buf.GetRecent(10)
	if len(all) != 5 {
		t.Errorf("len(GetRecent(10)) = %d, want 5", len(all))
	}

	// Get 0
	empty := buf.GetRecent(0)
	if len(empty) != 0 {
		t.Errorf("len(GetRecent(0)) = %d, want 0", len(empty))
	}
}

func TestOutputBufferLen(t *testing.T) {
	buf := NewOutputBuffer(10)

	if buf.Len() != 0 {
		t.Errorf("initial Len() = %d, want 0", buf.Len())
	}

	buf.Append(OutputLine{Text: "test"})
	if buf.Len() != 1 {
		t.Errorf("Len() after append = %d, want 1", buf.Len())
	}

	buf.Append(OutputLine{Text: "test2"})
	if buf.Len() != 2 {
		t.Errorf("Len() after second append = %d, want 2", buf.Len())
	}
}

func TestOutputBufferClear(t *testing.T) {
	buf := NewOutputBuffer(10)

	// Add some lines
	for i := 0; i < 5; i++ {
		buf.Append(OutputLine{Text: "test"})
	}

	if buf.Len() != 5 {
		t.Errorf("Len() before clear = %d, want 5", buf.Len())
	}

	buf.Clear()

	if buf.Len() != 0 {
		t.Errorf("Len() after clear = %d, want 0", buf.Len())
	}

	// Should be able to add lines after clear
	buf.Append(OutputLine{Text: "new"})
	if buf.Len() != 1 {
		t.Errorf("Len() after new append = %d, want 1", buf.Len())
	}
}

func TestOutputBufferConcurrency(t *testing.T) {
	buf := NewOutputBuffer(100)
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				buf.Append(OutputLine{
					Text:   "test",
					Stream: "stdout",
				})
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = buf.GetAll()
				_ = buf.GetRecent(10)
				_ = buf.Len()
			}
		}()
	}

	wg.Wait()

	// Should have some lines (may be less than 100 due to overflow)
	if buf.Len() == 0 {
		t.Error("Buffer should have lines after concurrent writes")
	}
	if buf.Len() > 100 {
		t.Errorf("Buffer should not exceed maxLines, got %d", buf.Len())
	}
}

func TestOutputLineFields(t *testing.T) {
	now := time.Now()
	line := OutputLine{
		Timestamp: now,
		Text:      "test output",
		Stream:    "stderr",
	}

	if line.Timestamp != now {
		t.Error("Timestamp not set correctly")
	}
	if line.Text != "test output" {
		t.Errorf("Text = %q, want %q", line.Text, "test output")
	}
	if line.Stream != "stderr" {
		t.Errorf("Stream = %q, want %q", line.Stream, "stderr")
	}
}
