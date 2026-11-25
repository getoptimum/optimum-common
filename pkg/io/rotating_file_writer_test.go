package io

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRotatingFileWriter(t *testing.T) {
	t.Parallel()

	t.Run("creates file and directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "logs", "test.log")

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close() //nolint:errcheck

		// Verify file exists
		_, err = os.Stat(filePath)
		require.NoError(t, err)
	})

	t.Run("uses default max size when zero", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		writer, err := NewRotatingFileWriter(filePath, 0)
		require.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close() //nolint:errcheck

		require.Equal(t, int64(DefaultMaxFileSize), writer.maxSize)
	})

	t.Run("uses default max size when negative", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		writer, err := NewRotatingFileWriter(filePath, -1)
		require.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close() //nolint:errcheck

		require.Equal(t, int64(DefaultMaxFileSize), writer.maxSize)
	})

	t.Run("preserves existing file size", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		// Create file with some content
		err := os.WriteFile(filePath, []byte("existing content"), 0644) //nolint:gosec // test file permissions
		require.NoError(t, err)

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close() //nolint:errcheck

		require.Equal(t, int64(16), writer.bytesWritten) // "existing content" = 16 bytes
	})
}

func TestRotatingFileWriter_Write(t *testing.T) {
	t.Parallel()

	t.Run("writes to file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)
		defer writer.Close() //nolint:errcheck

		data := []byte("test message")
		n, err := writer.Write(data)
		require.NoError(t, err)
		require.Equal(t, len(data), n)

		// Verify content
		content, err := os.ReadFile(filePath) //nolint:gosec // test file path
		require.NoError(t, err)
		require.Equal(t, data, content)
	})

	t.Run("appends to existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		// Create file with initial content
		err := os.WriteFile(filePath, []byte("initial "), 0644) //nolint:gosec // test file permissions
		require.NoError(t, err)

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)
		defer writer.Close() //nolint:errcheck

		data := []byte("appended")
		n, err := writer.Write(data)
		require.NoError(t, err)
		require.Equal(t, len(data), n)

		// Verify content
		content, err := os.ReadFile(filePath) //nolint:gosec // test file path
		require.NoError(t, err)
		require.Equal(t, []byte("initial appended"), content)
	})

	t.Run("rotates when max size exceeded", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		// Use small max size for testing (1KB)
		maxSize := int64(1024)
		writer, err := NewRotatingFileWriter(filePath, maxSize)
		require.NoError(t, err)
		defer writer.Close() //nolint:errcheck

		// Write data that exceeds max size
		largeData := make([]byte, maxSize+100)
		for i := range largeData {
			largeData[i] = byte('a')
		}

		n, err := writer.Write(largeData)
		require.NoError(t, err)
		require.Equal(t, len(largeData), n)

		// Verify rotation occurred
		rotatedFile := filePath + ".1"
		_, err = os.Stat(rotatedFile)
		require.NoError(t, err, "rotated file should exist")

		// Verify new file exists and is smaller
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		require.Less(t, info.Size(), maxSize)
	})

	t.Run("handles multiple rotations", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		// Use very small max size for testing (100 bytes)
		maxSize := int64(100)
		writer, err := NewRotatingFileWriter(filePath, maxSize)
		require.NoError(t, err)
		defer writer.Close() //nolint:errcheck

		// Write multiple times to trigger rotations
		for i := range 3 {
			data := make([]byte, maxSize+10)
			for j := range data {
				data[j] = byte('a' + i)
			}
			_, err := writer.Write(data)
			require.NoError(t, err)
		}

		// Verify rotated files exist
		for i := 1; i <= 3; i++ {
			rotatedFile := fmt.Sprintf("%s.%d", filePath, i)
			_, err := os.Stat(rotatedFile)
			require.NoError(t, err, "rotated file %d should exist", i)
		}
	})
}

func TestRotatingFileWriter_Close(t *testing.T) {
	t.Parallel()

	t.Run("closes file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		// Verify file is closed by checking we can't write
		_, err = writer.Write([]byte("test"))
		require.Error(t, err)
	})

	t.Run("handles double close", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		// Second close should not error
		err = writer.Close()
		require.NoError(t, err)
	})
}

func TestRotatingFileWriter_KeepsLogWhenRenameFails(t *testing.T) {
	t.Parallel()
	filePath := filepath.Join(t.TempDir(), "test.log")

	// Given every rotation slot is a non-empty directory, so each rename fails.
	for i := 1; i <= MaxRotatedFiles; i++ {
		slot := fmt.Sprintf("%s.%d", filePath, i)
		require.NoError(t, os.MkdirAll(filepath.Join(slot, "keep"), 0o755))
	}

	writer, err := NewRotatingFileWriter(filePath, 10)
	require.NoError(t, err)
	defer writer.Close() //nolint:errcheck

	// When a write crosses the size limit.
	payload := []byte("0123456789ABC")
	_, err = writer.Write(payload)
	require.NoError(t, err)

	// Then the live log is not truncated just because it could not be rotated.
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, string(payload), string(data))

	// When another write follows while rotation is still blocked.
	second := []byte("DEF")
	n, err := writer.Write(second)

	// Then the write succeeds and is appended to the live log.
	require.NoError(t, err)
	require.Equal(t, len(second), n)
	data, err = os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, string(payload)+string(second), string(data))
}

func TestRotatingFileWriter_RecoversAfterFailedReplacement(t *testing.T) {
	t.Parallel()
	filePath := filepath.Join(t.TempDir(), "test.log")

	writer, err := NewRotatingFileWriter(filePath, 10)
	require.NoError(t, err)

	// Given a rotation that moved the log but could not create the replacement.
	writer.mu.Lock()
	require.NoError(t, writer.current.Close())
	writer.current = nil
	writer.bytesWritten = writer.maxSize
	writer.mu.Unlock()

	// When the next write arrives after the filesystem has recovered.
	n, err := writer.Write([]byte("after"))

	// Then the rotation is retried and the write succeeds.
	require.NoError(t, err)
	require.Equal(t, 5, n)
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, "after", string(data))

	// And an explicit close still rejects later writes.
	require.NoError(t, writer.Close())
	_, err = writer.Write([]byte("x"))
	require.Error(t, err)
}
