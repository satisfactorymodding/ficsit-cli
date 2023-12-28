package cache

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/viper"
)

const IconFilename = "Resources/Icon128.png" // This is the path UE expects for the icon

type File struct {
	Icon         *string
	ModReference string
	Hash         string
	Plugin       UPlugin
	Size         int64
}

var loadedCache *xsync.MapOf[string, []File]

func GetCache() (*xsync.MapOf[string, []File], error) {
	if loadedCache != nil {
		return loadedCache, nil
	}
	return LoadCache()
}

func GetCacheMod(mod string) ([]File, error) {
	cache, err := GetCache()
	if err != nil {
		return nil, err
	}
	value, _ := cache.Load(mod)
	return value, nil
}

func LoadCache() (*xsync.MapOf[string, []File], error) {
	loadedCache = xsync.NewMapOf[string, []File]()
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if _, err := os.Stat(downloadCache); os.IsNotExist(err) {
		return loadedCache, nil
	}

	items, err := os.ReadDir(downloadCache)
	if err != nil {
		return nil, fmt.Errorf("failed reading download cache: %w", err)
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}
		if item.Name() == integrityFilename {
			continue
		}

		_, err = addFileToCache(item.Name())
		if err != nil {
			slog.Error("failed to add file to cache", slog.String("file", item.Name()), slog.Any("err", err))
		}
	}
	return loadedCache, nil
}

func addFileToCache(filename string) (*File, error) {
	cacheFile, err := readCacheFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	loadedCache.Compute(cacheFile.ModReference, func(oldValue []File, _ bool) ([]File, bool) {
		return append(oldValue, *cacheFile), false
	})

	return cacheFile, nil
}

func readCacheFile(filename string) (*File, error) {
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	path := filepath.Join(downloadCache, filename)
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	zipFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer zipFile.Close()

	size := stat.Size()
	reader, err := zip.NewReader(zipFile, size)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip: %w", err)
	}

	var upluginFile *zip.File
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".uplugin") {
			upluginFile = file
			break
		}
	}
	if upluginFile == nil {
		return nil, errors.New("no uplugin file found in zip")
	}

	upluginReader, err := upluginFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uplugin file: %w", err)
	}

	var uplugin UPlugin
	data, err := io.ReadAll(upluginReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read uplugin file: %w", err)
	}
	if err := json.Unmarshal(data, &uplugin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal uplugin file: %w", err)
	}

	modReference := strings.TrimSuffix(upluginFile.Name, ".uplugin")

	hash, err := getFileHash(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get file hash: %w", err)
	}

	var iconFile *zip.File
	for _, file := range reader.File {
		if file.Name == IconFilename {
			iconFile = file
			break
		}
	}
	var icon *string
	if iconFile != nil {
		iconReader, err := iconFile.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open icon file: %w", err)
		}
		defer iconReader.Close()

		data, err := io.ReadAll(iconReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read icon file: %w", err)
		}
		iconData := base64.StdEncoding.EncodeToString(data)
		icon = &iconData
	}

	return &File{
		ModReference: modReference,
		Hash:         hash,
		Size:         size,
		Icon:         icon,
		Plugin:       uplugin,
	}, nil
}
