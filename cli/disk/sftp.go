package disk

import (
	"io"
)

var _ Disk = (*sftpDisk)(nil)

type sftpDisk struct {
	path string
}

func newSFTP(path string) (Disk, error) {
	return sftpDisk{path: path}, nil
}

func (l sftpDisk) Exists(path string) error { //nolint
	panic("implement me")
}

func (l sftpDisk) Read(path string) ([]byte, error) { //nolint
	panic("implement me")
}

func (l sftpDisk) Write(path string, data []byte) error { //nolint
	panic("implement me")
}

func (l sftpDisk) Remove(path string) error { //nolint
	panic("implement me")
}

func (l sftpDisk) MkDir(path string) error { //nolint
	panic("implement me")
}

func (l sftpDisk) ReadDir(path string) ([]Entry, error) { //nolint
	panic("implement me")
}

func (l sftpDisk) IsNotExist(err error) bool { //nolint
	panic("implement me")
}

func (l sftpDisk) IsExist(err error) bool { //nolint
	panic("implement me")
}

func (l sftpDisk) Open(path string, flag int) (io.WriteCloser, error) { //nolint
	panic("implement me")
}
