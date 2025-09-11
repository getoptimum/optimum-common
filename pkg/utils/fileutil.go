package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"os"
	"path/filepath"
	"slices"
)

// CleanupFolders ensures that each directory in dirs exists and removes all entries except in ignoreFiles
func CleanupFolders(dirs []string, ignoreFiles ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("cannot create directory %q: %w", dir, err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("cannot read directory %q: %w", dir, err)
		}
		for _, entry := range entries {
			if slices.Contains(ignoreFiles, entry.Name()) {
				continue
			}
			entryPath := filepath.Join(dir, entry.Name())
			if err = os.RemoveAll(entryPath); err != nil {
				return fmt.Errorf("cannot remove %q: %w", entryPath, err)
			}
		}
	}
	return nil
}

// GetTracesDir returns path to store trace data for given protocol
func GetTracesDir(dataDir string, protocol fmt.Stringer) string {
	return filepath.Join(dataDir, "optimum", protocol.String(), "traces")
}

// GetPProfDir returns path to store pprof profiles for given protocol
func GetPProfDir(dataDir string, protocol fmt.Stringer) string {
	return filepath.Join(dataDir, "optimum", protocol.String(), "pprof")
}

// GetAuthUsageDir returns path to store auth usage data
func GetAuthUsageDir(dataDir string) string {
	return filepath.Join(dataDir, "optimum", "auth")
}

// AtomicallySaveToFile saves given data to the given file atomically
// Appends a CRC64 checksum to the data before writing
// File is either fully updated or not updated at all -> done by writing to a temporary file and renaming after
func AtomicallySaveToFile(fileName string, data []byte) error {
	checkSum := crc64.New(crc64.MakeTable(crc64.ISO))
	if _, err := checkSum.Write(data); err != nil {
		return fmt.Errorf("cannot calculate checksum: %w", err)
	}

	checkSumBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(checkSumBytes, checkSum.Sum64())
	resultData := make([]byte, 0, len(data)+len(checkSumBytes))
	resultData = append(resultData, checkSumBytes...)
	resultData = append(resultData, data...)

	dir, file := filepath.Split(fileName)
	if dir == "" {
		dir = "."
	}

	f, err := os.CreateTemp(dir, file)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	defer func() {
		if err != nil {
			_ = os.Remove(f.Name())
		}
	}()
	defer f.Close() //nolint:errcheck

	name := f.Name()
	if _, err = io.Copy(f, bytes.NewReader(resultData)); err != nil {
		return fmt.Errorf("cannot write data to tempfile %q: %w", name, err)
	}
	if err = f.Sync(); err != nil {
		return fmt.Errorf("can't flush tempfile %q: %w", name, err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("can't close tempfile %q: %w", name, err)
	}

	destInfo, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		// no original file
	} else if err != nil {
		return err
	} else {
		sourceInfo, errS := os.Stat(name)
		if errS != nil {
			return errS
		}
		if sourceInfo.Mode() != destInfo.Mode() {
			if err = os.Chmod(name, destInfo.Mode()); err != nil {
				return fmt.Errorf("can't set filemode on tempfile %q: %w", name, err)
			}
		}
	}
	if err = os.Rename(name, fileName); err != nil {
		return fmt.Errorf("cannot replace %q with tempfile %q: %w", fileName, name, err)
	}
	return nil
}

// LoadFromFile loads data from the given file and verifies the checksum
// Returns data w/o checksum
func LoadFromFile(path string) (data []byte, err error) {
	path = filepath.Clean(path)
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return nil, os.ErrNotExist
	}
	// #nosec G304 -- path is caller-controlled within the application; sanitized above
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer r.Close() //nolint:errcheck

	data, err = io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if len(data) < 8 {
		return nil, errors.New("file is too short")
	}
	fileCrc := binary.LittleEndian.Uint64(data[:8])
	dataCrc := crc64.Checksum(data[8:], crc64.MakeTable(crc64.ISO))
	if fileCrc != dataCrc {
		return nil, fmt.Errorf("checksum mismatch: %x != %x", fileCrc, dataCrc)
	}
	return data[8:], nil
}
