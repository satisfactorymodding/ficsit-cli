package disk

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
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
		return nil, fmt.Errorf("failed to parse ftp url: %w", err)
	}

	c, failedHidden, _ := testFTP(u, ftp.DialWithTimeout(time.Second*5), ftp.DialWithForceListHidden(true))
	if failedHidden {
		c, _, err = testFTP(u, ftp.DialWithTimeout(time.Second*5))
		if err != nil {
			return nil, err
		}
	}

	slog.Info("logged into ftp", slog.String("url", path), slog.Bool("hidden-files", !failedHidden))

	return &ftpDisk{
		path:   u.Path,
		client: c,
	}, nil
}

func testFTP(u *url.URL, options ...ftp.DialOption) (*ftp.ServerConn, bool, error) {
	c, err := ftp.Dial(u.Host, options...)
	if err != nil {
		return nil, false, fmt.Errorf("failed to dial host %s: %w", u.Host, err)
	}

	password, _ := u.User.Password()
	if err := c.Login(u.User.Username(), password); err != nil {
		return nil, false, fmt.Errorf("failed to login: %w", err)
	}

	_, err = c.List("/")
	if err != nil {
		return nil, true, fmt.Errorf("failed listing dir: %w", err)
	}

	return c, false, nil
}

func (l *ftpDisk) Exists(path string) (bool, error) {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	slog.Debug("checking if file exists", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	list, err := l.client.List(clean(filepath.Dir(path)))
	if err != nil {
		return false, fmt.Errorf("failed listing directory: %w", err)
	}

	found := false
	for _, entry := range list {
		if entry.Name == clean(filepath.Base(path)) {
			found = true
			break
		}
	}

	return found, nil
}

func (l *ftpDisk) Read(path string) ([]byte, error) {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	slog.Debug("reading file", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	f, err := l.client.Retr(clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve path: %w", err)
	}

	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func (l *ftpDisk) Write(path string, data []byte) error {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	slog.Debug("writing to file", slog.String("path", clean(path)), slog.String("schema", "ftp"))
	if err := l.client.Stor(clean(path), bytes.NewReader(data)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (l *ftpDisk) Remove(path string) error {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	slog.Debug("going to root directory", slog.String("schema", "ftp"))
	err := l.client.ChangeDir("/")
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	slog.Debug("deleting path", slog.String("path", clean(path)), slog.String("schema", "ftp"))
	if err := l.client.Delete(clean(path)); err != nil {
		if err := l.client.RemoveDirRecur(clean(path)); err != nil {
			return fmt.Errorf("failed to delete path: %w", err)
		}
	}

	return nil
}

func (l *ftpDisk) MkDir(path string) error {
	l.stepLock.Lock()
	defer l.stepLock.Unlock()

	slog.Debug("going to root directory", slog.String("schema", "ftp"))
	err := l.client.ChangeDir("/")
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	split := strings.Split(clean(path)[1:], "/")
	for _, s := range split {
		dir, err := l.ReadDirLock("", false)
		if err != nil {
			return err
		}

		currentDir, _ := l.client.CurrentDir()

		foundDir := false
		for _, entry := range dir {
			if entry.IsDir() && entry.Name() == s {
				foundDir = true
				break
			}
		}

		if !foundDir {
			slog.Debug("making directory", slog.String("dir", s), slog.String("cwd", currentDir), slog.String("schema", "ftp"))
			if err := l.client.MakeDir(s); err != nil {
				return fmt.Errorf("failed to make directory: %w", err)
			}
		}

		slog.Debug("entering directory", slog.String("dir", s), slog.String("cwd", currentDir), slog.String("schema", "ftp"))
		if err := l.client.ChangeDir(s); err != nil {
			return fmt.Errorf("failed to enter directory: %w", err)
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

	slog.Debug("reading directory", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	dir, err := l.client.List(clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to list files in directory: %w", err)
	}

	entries := make([]Entry, len(dir))
	for i, entry := range dir {
		entries[i] = ftpEntry{
			Entry: entry,
		}
	}

	return entries, nil
}

func (l *ftpDisk) Open(path string, _ int) (io.WriteCloser, error) {
	reader, writer := io.Pipe()

	slog.Debug("opening for writing", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	go func() {
		l.stepLock.Lock()
		defer l.stepLock.Unlock()

		err := l.client.Stor(clean(path), reader)
		if err != nil {
			slog.Error("failed to store file", slog.Any("err", err))
		}
		slog.Debug("write success", slog.String("path", clean(path)), slog.String("schema", "ftp"))
	}()

	return writer, nil
}
