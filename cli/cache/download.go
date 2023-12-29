package cache

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/utils"
)

type downloadGroup struct {
	err     error
	wait    chan bool
	hash    string
	updates []chan<- utils.GenericProgress
	size    int64
}

var downloadSync = *xsync.NewMapOf[string, *downloadGroup]()

func DownloadOrCache(cacheKey string, hash string, url string, updates chan<- utils.GenericProgress, downloadSemaphore chan int) (*os.File, int64, error) {
	group, loaded := downloadSync.LoadOrCompute(cacheKey, func() *downloadGroup {
		return &downloadGroup{
			hash:    hash,
			updates: make([]chan<- utils.GenericProgress, 0),
			wait:    make(chan bool),
		}
	})

	if updates != nil {
		_, _ = downloadSync.Compute(cacheKey, func(oldValue *downloadGroup, loaded bool) (*downloadGroup, bool) {
			oldValue.updates = append(oldValue.updates, updates)
			return oldValue, false
		})
	}

	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if err := os.MkdirAll(downloadCache, 0o777); err != nil {
		if !os.IsExist(err) {
			return nil, 0, fmt.Errorf("failed creating download cache: %w", err)
		}
	}

	location := filepath.Join(downloadCache, cacheKey)

	if loaded {
		if group.hash != hash {
			return nil, 0, errors.New("hash mismatch in download group")
		}

		<-group.wait

		if group.err != nil {
			return nil, 0, group.err
		}

		f, err := os.Open(location)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to open file: %s: %w", location, err)
		}

		return f, group.size, nil
	}

	defer downloadSync.Delete(cacheKey)

	upstreamUpdates := make(chan utils.GenericProgress)
	defer close(upstreamUpdates)

	upstreamWaiter := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

	outer:
		for {
			select {
			case update, ok := <-upstreamUpdates:
				if !ok {
					break outer
				}

				for _, u := range group.updates {
					u <- update
				}
			case <-upstreamWaiter:
				break outer
			}
		}
	}()

	var size int64

	err := retry.Do(func() error {
		var err error
		size, err = downloadInternal(cacheKey, location, hash, url, upstreamUpdates, downloadSemaphore)
		if err != nil {
			return fmt.Errorf("internal download error: %w", err)
		}
		return nil
	},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(n uint, err error) {
			if n > 0 {
				slog.Info("retrying download", slog.Uint64("n", uint64(n)), slog.String("cacheKey", cacheKey))
			}
		}),
	)
	if err != nil {
		group.err = err
		close(group.wait)
		return nil, 0, err // nolint
	}

	close(upstreamWaiter)
	wg.Wait()

	group.size = size
	close(group.wait)

	f, err := os.Open(location)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %s: %w", location, err)
	}

	return f, size, nil
}

func downloadInternal(cacheKey string, location string, hash string, url string, updates chan<- utils.GenericProgress, downloadSemaphore chan int) (int64, error) {
	stat, err := os.Stat(location)
	if err == nil {
		matches, err := compareHash(hash, location)
		if err != nil {
			return 0, err
		}

		if matches {
			return stat.Size(), nil
		}

		if err := os.Remove(location); err != nil {
			return 0, fmt.Errorf("failed to delete file: %s: %w", location, err)
		}
	} else if !os.IsNotExist(err) {
		return 0, fmt.Errorf("failed to stat file: %s: %w", location, err)
	}

	if updates != nil {
		headResp, err := http.Head(url)
		if err != nil {
			return 0, fmt.Errorf("failed to head: %s: %w", url, err)
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
		return 0, fmt.Errorf("failed creating file at: %s: %w", location, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch: %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status: %s on url: %s", resp.Status, url)
	}

	progresser := &utils.Progresser{
		Total:   resp.ContentLength,
		Updates: updates,
	}

	_, err = io.Copy(io.MultiWriter(out, progresser), resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed writing file to disk: %w", err)
	}

	_ = out.Sync()

	if updates != nil {
		updates <- utils.GenericProgress{Completed: resp.ContentLength, Total: resp.ContentLength}
	}

	_, err = addFileToCache(cacheKey)
	if err != nil {
		return 0, fmt.Errorf("failed to add file to cache: %w", err)
	}

	return resp.ContentLength, nil
}

func compareHash(hash string, location string) (bool, error) {
	existingHash := ""

	if hash != "" {
		f, err := os.Open(location)
		if err != nil {
			return false, fmt.Errorf("failed to open file: %s: %w", location, err)
		}
		defer f.Close()

		existingHash, err = utils.SHA256Data(f)
		if err != nil {
			return false, fmt.Errorf("could not compute hash for file: %s: %w", location, err)
		}
	}

	return hash == existingHash, nil
}
