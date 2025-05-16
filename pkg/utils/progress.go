package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressTracker manages progress bars for different operations
type ProgressTracker struct {
	mu    sync.Mutex
	bars  map[string]*progressbar.ProgressBar
	total int64
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		bars: make(map[string]*progressbar.ProgressBar),
	}
}

// AddProgressBar adds a new progress bar for an operation
func (pt *ProgressTracker) AddProgressBar(name string, total int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(name),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintf(os.Stderr, "\nCompleted: %s\n", name)
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	pt.bars[name] = bar
}

// Increment increases the progress of a named operation
func (pt *ProgressTracker) Increment(name string, n int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if bar, exists := pt.bars[name]; exists {
		bar.Add64(n)
	}
}

// Finish marks a progress bar as complete
func (pt *ProgressTracker) Finish(name string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if bar, exists := pt.bars[name]; exists {
		bar.Finish()
		delete(pt.bars, name)
		fmt.Fprintf(os.Stderr, "\nFinished: %s\n", name)
	}
}

// ProgressWriter wraps an io.Writer to track progress
type ProgressWriter struct {
	writer  io.Writer
	tracker *ProgressTracker
	name    string
}

// NewProgressWriter creates a new progress writer
func NewProgressWriter(writer io.Writer, tracker *ProgressTracker, name string) *ProgressWriter {
	return &ProgressWriter{
		writer:  writer,
		tracker: tracker,
		name:    name,
	}
}

// Write implements io.Writer
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	if err == nil {
		pw.tracker.Increment(pw.name, int64(n))
	}
	return n, err
}
