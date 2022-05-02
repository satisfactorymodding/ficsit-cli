package utils

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func DownloadOrCache(cacheKey string, hash string, url string) (r io.ReaderAt, size int64, err error) {
	downloadCache := path.Join(viper.GetString("cache-dir"), "downloadCache")
	if err := os.MkdirAll(downloadCache, 0777); err != nil {
		if !os.IsExist(err) {
			return nil, 0, errors.Wrap(err, "failed creating download cache")
		}
	}

	location := path.Join(downloadCache, cacheKey)

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

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed writing file to disk")
	}

	f, err := os.Open(location)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to open file: "+location)
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

func ExtractMod(f io.ReaderAt, size int64, location string) error {
	if err := os.MkdirAll(location, 0777); err != nil {
		if !os.IsExist(err) {
			return errors.Wrap(err, "failed to create mod directory: "+location)
		}
	} else {
		if err := os.RemoveAll(location); err != nil {
			return errors.Wrap(err, "failed to remove directory: "+location)
		}

		if err := os.MkdirAll(location, 0777); err != nil {
			return errors.Wrap(err, "failed to create mod directory: "+location)
		}
	}

	reader, err := zip.NewReader(f, size)
	if err != nil {
		return errors.Wrap(err, "failed to read file as zip")
	}

	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			outFileLocation := path.Join(location, file.Name)

			if err := os.MkdirAll(path.Dir(outFileLocation), 0777); err != nil {
				return errors.Wrap(err, "failed to create mod directory: "+location)
			}

			outFile, err := os.OpenFile(outFileLocation, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				return errors.Wrap(err, "failed to write to file: "+location)
			}

			inFile, err := file.Open()
			if err != nil {
				return errors.Wrap(err, "failed to process mod zip")
			}

			if _, err := io.Copy(outFile, inFile); err != nil {
				return errors.Wrap(err, "failed to write to file: "+location)
			}
		}
	}

	return nil
}
