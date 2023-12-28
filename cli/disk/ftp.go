package disk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/puddle/v2"
	"github.com/jlaffaye/ftp"
)

// TODO Make configurable
const connectionCount = 5

var _ Disk = (*ftpDisk)(nil)

type ftpDisk struct {
	pool *puddle.Pool[*ftp.ServerConn]
	path string
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

	pool, err := puddle.NewPool(&puddle.Config[*ftp.ServerConn]{
		Constructor: func(ctx context.Context) (*ftp.ServerConn, error) {
			c, failedHidden, err := testFTP(u, ftp.DialWithTimeout(time.Second*5), ftp.DialWithForceListHidden(true))
			if failedHidden {
				c, _, err = testFTP(u, ftp.DialWithTimeout(time.Second*5))
				if err != nil {
					return nil, err
				}
			} else if err != nil {
				return nil, err
			}

			slog.Info("logged into ftp", slog.Bool("hidden-files", !failedHidden))

			return c, nil
		},
		MaxSize: connectionCount,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &ftpDisk{
		path: u.Path,
		pool: pool,
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
	res, err := l.acquire()
	if err != nil {
		return false, err
	}

	defer res.Release()

	slog.Debug("checking if file exists", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	split := strings.Split(clean(path)[1:], "/")
	for _, s := range split[:len(split)-1] {
		dir, err := l.readDirLock(res, "")
		if err != nil {
			return false, err
		}

		currentDir, _ := res.Value().CurrentDir()

		foundDir := false
		for _, entry := range dir {
			if entry.IsDir() && entry.Name() == s {
				foundDir = true
				break
			}
		}

		if !foundDir {
			return false, nil
		}

		slog.Debug("entering directory", slog.String("dir", s), slog.String("cwd", currentDir), slog.String("schema", "ftp"))
		if err := res.Value().ChangeDir(s); err != nil {
			return false, fmt.Errorf("failed to enter directory: %w", err)
		}
	}

	dir, err := l.readDirLock(res, "")
	if err != nil {
		return false, fmt.Errorf("failed listing directory: %w", err)
	}

	found := false
	for _, entry := range dir {
		if entry.Name() == clean(filepath.Base(path)) {
			found = true
			break
		}
	}

	return found, nil
}

func (l *ftpDisk) Read(path string) ([]byte, error) {
	res, err := l.acquire()
	if err != nil {
		return nil, err
	}

	defer res.Release()

	slog.Debug("reading file", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	f, err := res.Value().Retr(clean(path))
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
	res, err := l.acquire()
	if err != nil {
		return err
	}

	defer res.Release()

	slog.Debug("writing to file", slog.String("path", clean(path)), slog.String("schema", "ftp"))
	if err := res.Value().Stor(clean(path), bytes.NewReader(data)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (l *ftpDisk) Remove(path string) error {
	res, err := l.acquire()
	if err != nil {
		return err
	}

	defer res.Release()

	slog.Debug("deleting path", slog.String("path", clean(path)), slog.String("schema", "ftp"))
	if err := res.Value().Delete(clean(path)); err != nil {
		if err := res.Value().RemoveDirRecur(clean(path)); err != nil {
			return fmt.Errorf("failed to delete path: %w", err)
		}
	}

	return nil
}

func (l *ftpDisk) MkDir(path string) error {
	res, err := l.acquire()
	if err != nil {
		return err
	}

	defer res.Release()

	split := strings.Split(clean(path)[1:], "/")
	for _, s := range split {
		dir, err := l.readDirLock(res, "")
		if err != nil {
			return err
		}

		currentDir, _ := res.Value().CurrentDir()

		foundDir := false
		for _, entry := range dir {
			if entry.IsDir() && entry.Name() == s {
				foundDir = true
				break
			}
		}

		if !foundDir {
			slog.Debug("making directory", slog.String("dir", s), slog.String("cwd", currentDir), slog.String("schema", "ftp"))
			if err := res.Value().MakeDir(s); err != nil {
				return fmt.Errorf("failed to make directory: %w", err)
			}
		}

		slog.Debug("entering directory", slog.String("dir", s), slog.String("cwd", currentDir), slog.String("schema", "ftp"))
		if err := res.Value().ChangeDir(s); err != nil {
			return fmt.Errorf("failed to enter directory: %w", err)
		}
	}

	return nil
}

func (l *ftpDisk) ReadDir(path string) ([]Entry, error) {
	res, err := l.acquire()
	if err != nil {
		return nil, err
	}

	defer res.Release()

	return l.readDirLock(res, path)
}

func (l *ftpDisk) readDirLock(res *puddle.Resource[*ftp.ServerConn], path string) ([]Entry, error) {
	slog.Debug("reading directory", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	dir, err := res.Value().List(clean(path))
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
	res, err := l.acquire()
	if err != nil {
		return nil, err
	}

	reader, writer := io.Pipe()

	slog.Debug("opening for writing", slog.String("path", clean(path)), slog.String("schema", "ftp"))

	go func() {
		defer res.Release()

		err := res.Value().Stor(clean(path), reader)
		if err != nil {
			slog.Error("failed to store file", slog.Any("err", err))
		}
		slog.Debug("write success", slog.String("path", clean(path)), slog.String("schema", "ftp"))
	}()

	return writer, nil
}

func (l *ftpDisk) goHome(res *puddle.Resource[*ftp.ServerConn]) error {
	slog.Debug("going to root directory", slog.String("schema", "ftp"))

	err := res.Value().ChangeDir("/")
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	return nil
}

func (l *ftpDisk) acquire() (*puddle.Resource[*ftp.ServerConn], error) {
	res, err := l.pool.Acquire(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed acquiring connection: %w", err)
	}

	if err := l.goHome(res); err != nil {
		return nil, err
	}

	return res, nil
}
