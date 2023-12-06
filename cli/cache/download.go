package cache

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/utils"
)

type Progresser struct {
	io.Reader
	updates chan utils.GenericUpdate
	total   int64
	running int64
}

func (pt *Progresser) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.running += int64(n)

	if err == nil {
		if pt.updates != nil {
			select {
			case pt.updates <- utils.GenericUpdate{Progress: float64(pt.running) / float64(pt.total)}:
			default:
			}
		}
	}

	if err == io.EOF {
		return n, io.EOF
	}

	return n, errors.Wrap(err, "failed to read")
}

func DownloadOrCache(cacheKey string, hash string, url string, updates chan utils.GenericUpdate) (io.ReaderAt, int64, error) {
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if err := os.MkdirAll(downloadCache, 0o777); err != nil {
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

			existingHash, err = utils.SHA256Data(f)
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
	} else if !os.IsNotExist(err) {
		return nil, 0, errors.Wrap(err, "failed to stat file: "+location)
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
		case updates <- utils.GenericUpdate{Progress: 1}:
		default:
		}
	}

	_, err = addFileToCache(cacheKey)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to add file to cache")
	}

	return f, resp.ContentLength, nil
}
