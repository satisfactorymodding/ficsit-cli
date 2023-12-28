package disk

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var _ Disk = (*localDisk)(nil)

type localDisk struct {
	path string
}

type localEntry struct {
	os.DirEntry
}

func newLocal(path string) (Disk, error) {
	return localDisk{path: path}, nil
}

func (l localDisk) Exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed checking file existence: %w", err)
	}

	return true, nil
}

func (l localDisk) Read(path string) ([]byte, error) {
	return os.ReadFile(path) //nolint
}

func (l localDisk) Write(path string, data []byte) error {
	return os.WriteFile(path, data, 0o777) //nolint
}

func (l localDisk) Remove(path string) error {
	return os.RemoveAll(path) //nolint
}

func (l localDisk) MkDir(path string) error {
	return os.MkdirAll(path, 0o777) //nolint
}

func (l localDisk) ReadDir(path string) ([]Entry, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err //nolint
	}

	entries := make([]Entry, len(dir))
	for i, entry := range dir {
		entries[i] = localEntry{
			DirEntry: entry,
		}
	}

	return entries, nil
}

func (l localDisk) Open(path string, flag int) (io.WriteCloser, error) {
	return os.OpenFile(path, flag, 0o777) //nolint
}
