package cache

import (
	"archive/zip"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/satisfactorymodding/ficsit-cli/utils"
	"github.com/spf13/viper"
)

type CacheFile struct {
	ModReference string
	Hash         string
	Plugin       UPlugin
}

var loadedCache map[string][]CacheFile

func GetCache() (map[string][]CacheFile, error) {
	if loadedCache != nil {
		return loadedCache, nil
	}
	return LoadCache()
}

func GetCacheMod(mod string) ([]CacheFile, error) {
	cache, err := GetCache()
	if err != nil {
		return nil, err
	}
	return cache[mod], nil
}

func LoadCache() (map[string][]CacheFile, error) {
	loadedCache = map[string][]CacheFile{}
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if _, err := os.Stat(downloadCache); os.IsNotExist(err) {
		return map[string][]CacheFile{}, nil
	}

	items, err := os.ReadDir(downloadCache)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading download cache")
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}

		fullpath := filepath.Join(downloadCache, item.Name())
		_, err = addFileToCache(fullpath)
		if err != nil {
			log.Err(err).Str("file", fullpath).Msg("failed to add file to cache")
		}
	}
	return loadedCache, nil
}

func addFileToCache(path string) (*CacheFile, error) {
	cacheFile, err := readCacheFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cache file")
	}

	loadedCache[cacheFile.ModReference] = append(loadedCache[cacheFile.ModReference], *cacheFile)
	return cacheFile, nil
}

func readCacheFile(path string) (*CacheFile, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to stat file")
	}

	zipFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer zipFile.Close()

	reader, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return nil, errors.Wrap(err, "failed to read zip")
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
		return nil, errors.Wrap(err, "failed to open uplugin file")
	}

	var uplugin UPlugin
	data, err := io.ReadAll(upluginReader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read uplugin file")
	}
	if err := json.Unmarshal(data, &uplugin); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal uplugin file")
	}

	modReference := strings.TrimSuffix(upluginFile.Name, ".uplugin")

	hash, err := utils.SHA256Data(zipFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash uplugin file")
	}

	return &CacheFile{
		ModReference: modReference,
		Hash:         hash,
		Plugin:       uplugin,
	}, nil
}
