package disk

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var _ Disk = (*sftpDisk)(nil)

type sftpDisk struct {
	client *sftp.Client
	path   string
}

type sftpEntry struct {
	os.FileInfo
}

func (f sftpEntry) IsDir() bool {
	return f.FileInfo.IsDir()
}

func (f sftpEntry) Name() string {
	return f.FileInfo.Name()
}

func newSFTP(path string) (Disk, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sftp url: %w", err)
	}

	password, ok := u.User.Password()
	var auth []ssh.AuthMethod
	if ok {
		auth = append(auth, ssh.Password(password))
	}

	conn, err := ssh.Dial("tcp", u.Host, &ssh.ClientConfig{
		User: u.User.Username(),
		Auth: auth,

		// TODO Somehow use systems hosts file
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ssh server: %w", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create sftp client: %w", err)
	}

	return sftpDisk{
		path:   path,
		client: client,
	}, nil
}

func (l sftpDisk) Exists(path string) error { //nolint
	slog.Debug("checking if file exists", slog.String("path", path), slog.String("schema", "sftp"))

	_, err := l.client.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	return nil
}

func (l sftpDisk) Read(path string) ([]byte, error) {
	slog.Debug("reading file", slog.String("path", path), slog.String("schema", "sftp"))

	f, err := l.client.Open(path)
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

func (l sftpDisk) Write(path string, data []byte) error {
	slog.Debug("writing to file", slog.String("path", path), slog.String("schema", "sftp"))

	file, err := l.client.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer file.Close()

	if _, err = io.Copy(file, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (l sftpDisk) Remove(path string) error {
	slog.Debug("deleting path", slog.String("path", path), slog.String("schema", "sftp"))
	if err := l.client.Remove(path); err != nil {
		if err := l.client.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to delete path: %w", err)
		}
	}

	return nil
}

func (l sftpDisk) MkDir(path string) error {
	slog.Debug("making directory", slog.String("path", path), slog.String("schema", "sftp"))

	if err := l.client.MkdirAll(path); err != nil {
		return fmt.Errorf("failed to make directory: %w", err)
	}

	return nil
}

func (l sftpDisk) ReadDir(path string) ([]Entry, error) {
	dir, err := l.client.ReadDir(path)

	slog.Debug("reading directory", slog.String("path", path), slog.String("schema", "sftp"))

	if err != nil {
		return nil, fmt.Errorf("failed to list files in directory: %w", err)
	}

	entries := make([]Entry, len(dir))
	for i, entry := range dir {
		entries[i] = sftpEntry{
			FileInfo: entry,
		}
	}

	return entries, nil
}

func (l sftpDisk) IsNotExist(err error) bool {
	return err != nil && strings.Contains(err.Error(), "file does not exist")
}

func (l sftpDisk) IsExist(_ error) bool {
	return false
}

func (l sftpDisk) Open(path string, _ int) (io.WriteCloser, error) {
	slog.Debug("opening for writing", slog.String("path", path), slog.String("schema", "sftp"))

	f, err := l.client.Create(path)
	if err != nil {
		slog.Error("failed to open file", slog.Any("err", err))
	}

	return f, nil
}
