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

func (l sftpDisk) Exists(path string) error {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) Read(path string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) Write(path string, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) Remove(path string) error {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) MkDir(path string) error {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) ReadDir(path string) ([]Entry, error) {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) IsNotExist(err error) bool {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) IsExist(err error) bool {
	//TODO implement me
	panic("implement me")
}

func (l sftpDisk) Open(path string, flag int) (io.WriteCloser, error) {
	//TODO implement me
	panic("implement me")
}
