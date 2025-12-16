package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DailyFileWriter writes logs to files that rotate daily
type DailyFileWriter struct {
	directory   string
	filename    string
	file        *os.File
	mu          sync.Mutex
	currentDate string
}

// NewDailyFileWriter creates a new daily file writer
func NewDailyFileWriter(directory, filename string) (*DailyFileWriter, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	w := &DailyFileWriter{
		directory: directory,
		filename:  filename,
	}

	if err := w.rotateIfNeeded(); err != nil {
		return nil, err
	}

	return w, nil
}

// Write implements io.Writer interface
func (w *DailyFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if we need to rotate
	if err := w.rotateIfNeeded(); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

// Sync flushes the file
func (w *DailyFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Close closes the current file
func (w *DailyFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// rotateIfNeeded rotates the file if the date has changed
func (w *DailyFileWriter) rotateIfNeeded() error {
	today := time.Now().Format("2006-01-02")

	// If date hasn't changed and file is open, no need to rotate
	if w.currentDate == today && w.file != nil {
		return nil
	}

	// Close current file if open
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}

	// Open new file for today
	filename := fmt.Sprintf("%s-%s.log", w.filename, today)
	filepath := filepath.Join(w.directory, filename)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	w.file = file
	w.currentDate = today

	return nil
}

// GetCurrentFilePath returns the current log file path
func (w *DailyFileWriter) GetCurrentFilePath() string {
	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.log", w.filename, today)
	return filepath.Join(w.directory, filename)
}
