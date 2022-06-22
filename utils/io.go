package utils

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli/disk"
)

type Progresser struct {
	io.Reader
	total   int64
	running int64
	updates chan GenericUpdate
}

func (pt *Progresser) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.running += int64(n)

	if err == nil {
		if pt.updates != nil {
			select {
			case pt.updates <- GenericUpdate{Progress: float64(pt.running) / float64(pt.total)}:
			default:
			}
		}
	}

	if err == io.EOF {
		return n, io.EOF
	}

	return n, errors.Wrap(err, "failed to read")
}

type GenericUpdate struct {
	Progress     float64
	ModReference *string
}

func DownloadOrCache(cacheKey string, hash string, url string, updates chan GenericUpdate) (r io.ReaderAt, size int64, err error) {
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if err := os.MkdirAll(downloadCache, 0777); err != nil {
		if !os.IsExist(err) {
			return nil, 0, errors.Wrap(err, "failed creating download cache")
		}
	}

	location := filepath.Join(downloadCache, cacheKey)

	stat, err := os.Stat(location)
	if err == nil {
		existingHash := ""

		if hash != "" {
			f, err := os.Open(location)
			if err != nil {
				return nil, 0, errors.Wrap(err, "failed to open file: "+location)
			}

			existingHash, err = SHA256Data(f)
			if err != nil {
				return nil, 0, errors.Wrap(err, "could not compute hash for file: "+location)
			}
		}

		if hash == existingHash {
			f, err := os.Open(location)
			if err != nil {
				return nil, 0, errors.Wrap(err, "failed to open file: "+location)
			}

			return f, stat.Size(), nil
		}

		if err := os.Remove(location); err != nil {
			return nil, 0, errors.Wrap(err, "failed to delete file: "+location)
		}
	} else {
		if !os.IsNotExist(err) {
			return nil, 0, errors.Wrap(err, "failed to stat file: "+location)
		}
	}

	out, err := os.Create(location)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed creating file at: "+location)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to fetch: "+url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("bad status: %s on url: %s", resp.Status, url)
	}

	progresser := &Progresser{
		Reader:  resp.Body,
		total:   resp.ContentLength,
		updates: updates,
	}

	_, err = io.Copy(out, progresser)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed writing file to disk")
	}

	f, err := os.Open(location)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to open file: "+location)
	}

	if updates != nil {
		select {
		case updates <- GenericUpdate{Progress: 1}:
		default:
		}
	}

	return f, resp.ContentLength, nil
}

func SHA256Data(f io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errors.Wrap(err, "failed to compute hash")
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func ExtractMod(f io.ReaderAt, size int64, location string, hash string, updates chan GenericUpdate, d disk.Disk) error {
	hashFile := filepath.Join(location, ".smm")
	hashBytes, err := d.Read(hashFile)
	if err != nil {
		if !d.IsNotExist(err) {
			return errors.Wrap(err, "failed to read .smm mod hash file")
		}
	} else {
		if hash == string(hashBytes) {
			return nil
		}
	}

	if err := d.MkDir(location); err != nil {
		if !d.IsExist(err) {
			return errors.Wrap(err, "failed to create mod directory: "+location)
		}

		if err := d.Remove(location); err != nil {
			return errors.Wrap(err, "failed to remove directory: "+location)
		}

		if err := d.MkDir(location); err != nil {
			return errors.Wrap(err, "failed to create mod directory: "+location)
		}
	}

	reader, err := zip.NewReader(f, size)
	if err != nil {
		return errors.Wrap(err, "failed to read file as zip")
	}

	for i, file := range reader.File {
		if !file.FileInfo().IsDir() {
			outFileLocation := filepath.Join(location, file.Name)

			if err := d.MkDir(filepath.Dir(outFileLocation)); err != nil {
				return errors.Wrap(err, "failed to create mod directory: "+location)
			}

			if err := writeZipFile(outFileLocation, file, d); err != nil {
				return err
			}
		}

		if updates != nil {
			select {
			case updates <- GenericUpdate{Progress: float64(i) / float64(len(reader.File)-1)}:
			default:
			}
		}
	}

	if err := d.Write(hashFile, []byte(hash)); err != nil {
		return errors.Wrap(err, "failed to write .smm mod hash file")
	}

	if updates != nil {
		select {
		case updates <- GenericUpdate{Progress: 1}:
		default:
		}
	}

	return nil
}

func writeZipFile(outFileLocation string, file *zip.File, d disk.Disk) error {
	outFile, err := d.Open(outFileLocation, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return errors.Wrap(err, "failed to write to file: "+outFileLocation)
	}

	defer outFile.Close()

	inFile, err := file.Open()
	if err != nil {
		return errors.Wrap(err, "failed to process mod zip")
	}

	if _, err := io.Copy(outFile, inFile); err != nil {
		return errors.Wrap(err, "failed to write to file: "+outFileLocation)
	}

	return nil
}
