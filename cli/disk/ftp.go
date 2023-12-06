package disk

import (
	"bytes"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var _ Disk = (*ftpDisk)(nil)

type ftpDisk struct {
	client   *ftp.ServerConn
	path     string
	stepLock sync.Mutex
}

type ftpEntry struct {
	*ftp.Entry
}

func (f ftpEntry) IsDir() bool {
	return f.Entry.Type == ftp.EntryTypeFolder
}

func (f ftpEntry) Name() string {
	return f.Entry.Name
}

func newFTP(path string) (Disk, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ftp url")
	}

	c, err := ftp.Dial(u.Host, ftp.DialWithTimeout(time.Second*5))
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial host "+u.Host)
	}

	password, _ := u.User.Password()
	if err := c.Login(u.User.Username(), password); err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}

	log.Debug().Msg("logged into ftp")

	return &ftpDisk{
		path:   u.Path,
		client: c,
	}, nil
}

func (l *ftpDisk) Exists(path string) error {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	log.Debug().Str("path", path).Str("schema", "ftp").Msg("checking if file exists")
	_, err := l.client.FileSize(path)
	return errors.Wrap(err, "failed to check if file exists")
}

func (l *ftpDisk) Read(path string) ([]byte, error) {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	log.Debug().Str("path", path).Str("schema", "ftp").Msg("reading file")

	f, err := l.client.Retr(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve path")
	}

	defer f.Close()

	data, err := io.ReadAll(f)
	return data, errors.Wrap(err, "failed to read file")
}

func (l *ftpDisk) Write(path string, data []byte) error {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	log.Debug().Str("path", path).Str("schema", "ftp").Msg("writing to file")
	return errors.Wrap(l.client.Stor(path, bytes.NewReader(data)), "failed to write file")
}

func (l *ftpDisk) Remove(path string) error {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	log.Debug().Str("path", path).Str("schema", "ftp").Msg("deleting path")
	return errors.Wrap(l.client.Delete(path), "failed to delete path")
}

func (l *ftpDisk) MkDir(path string) error {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	log.Debug().Str("schema", "ftp").Msg("going to root directory")
	err := l.client.ChangeDir("/")
	if err != nil {
		return errors.Wrap(err, "failed to change directory")
	}

	split := strings.Split(path[1:], "/")
	for _, s := range split {
		dir, err := l.ReadDirLock("", false)
		if err != nil {
			return err
		}

		foundDir := false
		for _, entry := range dir {
			if entry.IsDir() && entry.Name() == s {
				foundDir = true
				break
			}
		}

		if !foundDir {
			log.Debug().Str("dir", s).Str("schema", "ftp").Msg("making directory")
			if err := l.client.MakeDir(s); err != nil {
				return errors.Wrap(err, "failed to make directory")
			}
		}

		log.Debug().Str("dir", s).Str("schema", "ftp").Msg("entering directory")
		if err := l.client.ChangeDir(s); err != nil {
			return errors.Wrap(err, "failed to enter directory")
		}
	}

	return nil
}

func (l *ftpDisk) ReadDir(path string) ([]Entry, error) {
	return l.ReadDirLock(path, true)
}

func (l *ftpDisk) ReadDirLock(path string, lock bool) ([]Entry, error) {
	if lock {
		l.stepLock.Lock()
		defer l.stepLock.Unlock()
	}

	log.Debug().Str("path", path).Str("schema", "ftp").Msg("reading directory")

	dir, err := l.client.List(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list files in directory")
	}

	entries := make([]Entry, len(dir))
	for i, entry := range dir {
		entries[i] = ftpEntry{
			Entry: entry,
		}
	}

	return entries, nil
}

func (l *ftpDisk) IsNotExist(err error) bool {
	return strings.Contains(err.Error(), "Could not get file") || strings.Contains(err.Error(), "Failed to open file")
}

func (l *ftpDisk) IsExist(err error) bool {
	return strings.Contains(err.Error(), "Create directory operation failed")
}

func (l *ftpDisk) Open(path string, _ int) (io.WriteCloser, error) {
	reader, writer := io.Pipe()

	log.Debug().Str("path", path).Str("schema", "ftp").Msg("opening for writing")

	go func() {
		l.stepLock.Lock()
		defer l.stepLock.Unlock()

		err := l.client.Stor(path, reader)
		if err != nil {
			log.Err(err).Msg("failed to store file")
		}
		log.Debug().Str("path", path).Str("schema", "ftp").Msg("write success")
	}()

	return writer, nil
}
