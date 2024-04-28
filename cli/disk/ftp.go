package disk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/textproto"
	"net/url"
	"path"
	"slices"
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

func (l *ftpDisk) existsWithLock(res *puddle.Resource[*ftp.ServerConn], p string) (bool, error) {
	slog.Debug("checking if file exists", slog.String("path", clean(p)), slog.String("schema", "ftp"))

	var protocolError *textproto.Error

	_, err := res.Value().GetEntry(clean(p))
	if err == nil {
		return true, nil
	}

	if errors.As(err, &protocolError) {
		switch protocolError.Code {
		case ftp.StatusFileUnavailable:
			return false, nil
		case ftp.StatusNotImplemented:
			// GetEntry uses MLST, which might not be supported by the server.
			// Even though in this case the error is not coming from the server,
			// the ftp library still returns it as a protocol error.
		default:
			// We won't handle any other kind of error, such as
			// * temporary errors (4xx) - should be retried after a while, so we won't deal with the delay
			// * connection errors (x2x) - can't really do anything about them
			// * authentication errors (x3x) - can't do anything about them
			return false, fmt.Errorf("failed to get path info: %w", err)
		}
	} else {
		// This is a non-protocol error, so we can't be sure what it means.
		return false, fmt.Errorf("failed to get path info: %w", err)
	}

	// In case MLST is not supported, we can try to LIST the target path.
	// We can be sure that List() will actually execute LIST and not MLSD,
	// since MLST was not supported in the previous step.
	entries, err := res.Value().List(clean(p))
	if err == nil {
		if len(entries) > 0 {
			// Some server implementations return an empty list for a nonexistent path,
			// so we cannot be sure that no error means a directory exists unless it also contains some items.
			// For files, when they exist, they will be listed as a single entry.
			// TODO: so far the servers (just one) this was happening on also listed . and .. for valid dirs, because it was using `LIST -a`. Is that behaviour consistent that we can rely on it?
			return true, nil
		}
	} else {
		if errors.As(err, &protocolError) {
			if protocolError.Code == ftp.StatusFileUnavailable {
				return false, nil
			}
		}
		// We won't handle any other kind of error, see above.
		return false, fmt.Errorf("failed to list path: %w", err)
	}

	// If we got here, either the path is an empty directory,
	// or it does not exist and the server is a weird implementation.

	// List the parent directory to determine if the path exists
	dir, err := l.readDirLock(res, path.Dir(clean(p)))
	if err == nil {
		found := false
		for _, entry := range dir {
			if entry.Name() == path.Base(clean(p)) {
				found = true
				break
			}
		}

		return found, nil
	}

	if errors.As(err, &protocolError) {
		if protocolError.Code == ftp.StatusFileUnavailable {
			return false, nil
		}
	}

	// We won't handle any other kind of error, see above.
	return false, fmt.Errorf("failed to list parent path: %w", err)
}

func (l *ftpDisk) Exists(p string) (bool, error) {
	res, err := l.acquire()
	if err != nil {
		return false, err
	}

	defer res.Release()

	return l.existsWithLock(res, p)
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

func (l *ftpDisk) MkDir(p string) error {
	res, err := l.acquire()
	if err != nil {
		return err
	}

	defer res.Release()

	lastExistingDir := clean(p)
	for lastExistingDir != "/" && lastExistingDir != "." {
		foundDir, err := l.existsWithLock(res, lastExistingDir)
		if err != nil {
			return err
		}

		if foundDir {
			break
		}

		lastExistingDir = path.Dir(lastExistingDir)
	}

	remainingDirs := clean(p)

	if lastExistingDir != "/" && lastExistingDir != "." {
		remainingDirs = strings.TrimPrefix(remainingDirs, lastExistingDir)
	}

	if len(remainingDirs) == 0 {
		// Already exists
		return nil
	}

	if err := res.Value().ChangeDir(lastExistingDir); err != nil {
		return fmt.Errorf("failed to enter directory: %w", err)
	}

	split := strings.Split(clean(remainingDirs)[1:], "/")
	for _, s := range split {
		slog.Debug("making directory", slog.String("dir", s), slog.String("cwd", lastExistingDir), slog.String("schema", "ftp"))
		if err := res.Value().MakeDir(s); err != nil {
			return fmt.Errorf("failed to make directory: %w", err)
		}

		slog.Debug("entering directory", slog.String("dir", s), slog.String("cwd", lastExistingDir), slog.String("schema", "ftp"))
		if err := res.Value().ChangeDir(s); err != nil {
			return fmt.Errorf("failed to enter directory: %w", err)
		}
		lastExistingDir = path.Join(lastExistingDir, s)
	}

	return nil
}

func (l *ftpDisk) ReadDir(path string) ([]Entry, error) {
	res, err := l.acquire()
	if err != nil {
		return nil, err
	}

	defer res.Release()

	entries, err := l.readDirLock(res, path)
	if err != nil {
		return nil, err
	}
	entries = slices.DeleteFunc(entries, func(i Entry) bool {
		return i.Name() == "." || i.Name() == ".."
	})
	return entries, nil
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
