package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// DefaultMaxFileSize is the default maximum file size before rotation (10MB)
	DefaultMaxFileSize = 10 * 1024 * 1024
	// MaxRotatedFiles is the maximum number of rotated files to keep (current + 5 rotated = 60MB total)
	MaxRotatedFiles = 5
)

// RotatingFileWriter is an io.Writer that rotates files when they exceed a size limit
type RotatingFileWriter struct {
	filePath     string
	maxSize      int64
	current      *os.File
	bytesWritten int64
	mu           sync.Mutex
}

// NewRotatingFileWriter creates a new RotatingFileWriter
// filePath: path to the log file (e.g., "/proxy/logs/proxy-debug.log" or "./logs/gateway-debug.log")
// maxSize: maximum file size in bytes before rotation (default: 10MB)
func NewRotatingFileWriter(filePath string, maxSize int64) (*RotatingFileWriter, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxFileSize
	}

	// Ensure we have the full path
	filePath = filepath.Join(filepath.Dir(filePath), filepath.Base(filePath))

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed create logs dir: %w", err)
	}

	// Open or create the file
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600) //nolint:gosec // log file permissions
	if err != nil {
		return nil, fmt.Errorf("failed open log file: %w", err)
	}

	// Get current file size
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed get file state: %w", err)
	}

	return &RotatingFileWriter{
		filePath:     filePath,
		maxSize:      maxSize,
		current:      f,
		bytesWritten: info.Size(),
	}, nil
}

// Write implements io.Writer
func (r *RotatingFileWriter) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.current == nil {
		return 0, fmt.Errorf("file writer is closed")
	}

	// Check if we need to rotate before writing
	if r.bytesWritten >= r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, fmt.Errorf("failed rotate log file: %w", err)
		}
	}

	// Write to file
	n, err = r.current.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed write to log file: %w", err)
	}

	r.bytesWritten += int64(n)

	// Check if we need to rotate after writing
	if r.bytesWritten >= r.maxSize {
		_ = r.rotate() // Ignore rotation errors, we've already written the data
	}

	return n, nil
}

// rotate rotates the log file
func (r *RotatingFileWriter) rotate() error {
	// Close current file
	if r.current != nil {
		_ = r.current.Close()
	}

	// Rotate existing files: file.log -> file.log.1 -> file.log.2, etc.
	// Remove the oldest file if it exists
	oldestPath := fmt.Sprintf("%s.%d", r.filePath, MaxRotatedFiles)
	if _, err := os.Stat(oldestPath); err == nil {
		_ = os.Remove(oldestPath)
	}

	// Shift rotated files
	for i := MaxRotatedFiles - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", r.filePath, i)
		newPath := fmt.Sprintf("%s.%d", r.filePath, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			_ = os.Rename(oldPath, newPath)
		}
	}

	// Move current file to .1
	if _, err := os.Stat(r.filePath); err == nil {
		_ = os.Rename(r.filePath, fmt.Sprintf("%s.1", r.filePath))
	}

	// Create new file
	f, err := os.OpenFile(r.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600) //nolint:gosec // log file permissions
	if err != nil {
		return fmt.Errorf("failed create new log file: %w", err)
	}

	r.current = f
	r.bytesWritten = 0
	return nil
}

// Close implements io.Closer
func (r *RotatingFileWriter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.current != nil {
		err := r.current.Close()
		r.current = nil
		return err
	}
	return nil
}
