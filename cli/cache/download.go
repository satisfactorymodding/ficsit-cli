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

func DownloadOrCache(cacheKey string, hash string, url string, updates chan<- utils.GenericProgress, downloadSemaphore chan int) (*os.File, int64, error) {
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
			defer f.Close()

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

	if updates != nil {
		headResp, err := http.Head(url)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to head: "+url)
		}
		defer headResp.Body.Close()
		updates <- utils.GenericProgress{Total: headResp.ContentLength}
	}

	if downloadSemaphore != nil {
		downloadSemaphore <- 1
		defer func() { <-downloadSemaphore }()
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

	progresser := &utils.Progresser{
		Reader:  resp.Body,
		Total:   resp.ContentLength,
		Updates: updates,
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
		updates <- utils.GenericProgress{Completed: resp.ContentLength, Total: resp.ContentLength}
	}

	_, err = addFileToCache(cacheKey)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to add file to cache")
	}

	return f, resp.ContentLength, nil
}
