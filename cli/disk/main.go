package disk

import (
	"io"
	"net/url"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Disk interface {
	Exists(path string) error
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Remove(path string) error
	MkDir(path string) error
	ReadDir(path string) ([]Entry, error)
	IsNotExist(err error) bool
	IsExist(err error) bool
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

	log.Info().Msg(path)
	log.Info().Msg(parsed.Scheme)
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
