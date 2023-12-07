package disk

import (
	"io"
	"net/url"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Disk interface {
	// Exists checks if the provided file or directory exists
	Exists(path string) error

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

	// IsNotExist returns true if provided error is a not-exist type error
	IsNotExist(err error) bool

	// IsExist returns true if provided error is a does-exist type error
	IsExist(err error) bool

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
		return nil, errors.Wrap(err, "failed to parse path")
	}

	switch parsed.Scheme {
	case "ftp":
		log.Info().Str("path", path).Msg("connecting to ftp")
		return newFTP(path)
	case "sftp":
		log.Info().Str("path", path).Msg("connecting to sftp")
		return newSFTP(path)
	}

	log.Info().Str("path", path).Msg("using local disk")
	return newLocal(path)
}
