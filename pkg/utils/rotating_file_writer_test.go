package utils

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
		defer writer.Close()

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
		defer writer.Close()

		require.Equal(t, int64(DefaultMaxFileSize), writer.maxSize)
	})

	t.Run("uses default max size when negative", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		writer, err := NewRotatingFileWriter(filePath, -1)
		require.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close()

		require.Equal(t, int64(DefaultMaxFileSize), writer.maxSize)
	})

	t.Run("preserves existing file size", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		// Create file with some content
		err := os.WriteFile(filePath, []byte("existing content"), 0644)
		require.NoError(t, err)

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)
		require.NotNil(t, writer)
		defer writer.Close()

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
		defer writer.Close()

		data := []byte("test message")
		n, err := writer.Write(data)
		require.NoError(t, err)
		require.Equal(t, len(data), n)

		// Verify content
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		require.Equal(t, data, content)
	})

	t.Run("appends to existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.log")

		// Create file with initial content
		err := os.WriteFile(filePath, []byte("initial "), 0644)
		require.NoError(t, err)

		writer, err := NewRotatingFileWriter(filePath, DefaultMaxFileSize)
		require.NoError(t, err)
		defer writer.Close()

		data := []byte("appended")
		n, err := writer.Write(data)
		require.NoError(t, err)
		require.Equal(t, len(data), n)

		// Verify content
		content, err := os.ReadFile(filePath)
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
		defer writer.Close()

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
		defer writer.Close()

		// Write multiple times to trigger rotations
		for i := 0; i < 3; i++ {
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
