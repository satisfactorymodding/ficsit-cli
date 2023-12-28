package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/utils"
)

type hashInfo struct {
	Modified time.Time
	Hash     string
	Size     int64
}

var hashCache *xsync.MapOf[string, hashInfo]

var integrityFilename = ".integrity"

func getFileHash(file string) (string, error) {
	if hashCache == nil {
		loadHashCache()
	}
	cachedHash, ok := hashCache.Load(file)
	if !ok {
		return cacheFileHash(file)
	}
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	stat, err := os.Stat(filepath.Join(downloadCache, file))
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	if stat.Size() != cachedHash.Size || stat.ModTime() != cachedHash.Modified {
		return cacheFileHash(file)
	}
	return cachedHash.Hash, nil
}

func cacheFileHash(file string) (string, error) {
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	stat, err := os.Stat(filepath.Join(downloadCache, file))
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	f, err := os.Open(filepath.Join(downloadCache, file))
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	hash, err := utils.SHA256Data(f)
	if err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}
	hashCache.Store(file, hashInfo{
		Hash:     hash,
		Size:     stat.Size(),
		Modified: stat.ModTime(),
	})
	saveHashCache()
	return hash, nil
}

func loadHashCache() {
	hashCache = xsync.NewMapOf[string, hashInfo]()
	cacheFile := filepath.Join(viper.GetString("cache-dir"), "downloadCache", integrityFilename)
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return
	}
	f, err := os.Open(cacheFile)
	if err != nil {
		slog.Warn("failed to open hash cache, recreating", slog.Any("err", err))
		return
	}
	defer f.Close()

	hashCacheJSON, err := io.ReadAll(f)
	if err != nil {
		slog.Warn("failed to read hash cache, recreating", slog.Any("err", err))
		return
	}

	if err := json.Unmarshal(hashCacheJSON, &hashCache); err != nil {
		slog.Warn("failed to unmarshal hash cache, recreating", slog.Any("err", err))
		return
	}
}

func saveHashCache() {
	cacheFile := filepath.Join(viper.GetString("cache-dir"), "downloadCache", integrityFilename)
	plainCache := make(map[string]hashInfo, hashCache.Size())
	hashCache.Range(func(k string, v hashInfo) bool {
		plainCache[k] = v
		return true
	})
	hashCacheJSON, err := json.Marshal(plainCache)
	if err != nil {
		slog.Warn("failed to marshal hash cache", slog.Any("err", err))
		return
	}

	if err := os.WriteFile(cacheFile, hashCacheJSON, 0o755); err != nil {
		slog.Warn("failed to write hash cache", slog.Any("err", err))
		return
	}
}
