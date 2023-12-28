package disk

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"path/filepath"
)

type Disk interface {
	// Exists checks if the provided file or directory exists
	Exists(path string) (bool, error)

	// Read returns the entire file as a byte buffer
	//
	// Returns error if provided path is not a file
	Read(path string) ([]byte, error)

	// Write writes provided byte buffer to the path
	Write(path string, data []byte) error

	// Remove deletes the provided file or directory recursively
	Remove(path string) error

	// MkDir creates the provided directory recursively
	MkDir(path string) error

	// ReadDir returns all entries within the directory
	//
	// Returns error if provided path is not a directory
	ReadDir(path string) ([]Entry, error)

	// Open opens provided path for writing
	Open(path string, flag int) (io.WriteCloser, error)
}

type Entry interface {
	IsDir() bool
	Name() string
}

func FromPath(path string) (Disk, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %w", err)
	}

	switch parsed.Scheme {
	case "ftp":
		slog.Info("connecting to ftp")
		return newFTP(path)
	case "sftp":
		slog.Info("connecting to sftp")
		return newSFTP(path)
	}

	slog.Info("using local disk", slog.String("path", path))
	return newLocal(path)
}

// clean returns a unix-style path
func clean(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}
