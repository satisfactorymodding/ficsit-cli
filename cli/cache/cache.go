package cache

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

var loadedCache map[string][]File

func GetCache() (map[string][]File, error) {
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
	return cache[mod], nil
}

func LoadCache() (map[string][]File, error) {
	loadedCache = map[string][]File{}
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if _, err := os.Stat(downloadCache); os.IsNotExist(err) {
		return map[string][]File{}, nil
	}

	items, err := os.ReadDir(downloadCache)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading download cache")
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
			log.Err(err).Str("file", item.Name()).Msg("failed to add file to cache")
		}
	}
	return loadedCache, nil
}

func addFileToCache(filename string) (*File, error) {
	cacheFile, err := readCacheFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cache file")
	}

	loadedCache[cacheFile.ModReference] = append(loadedCache[cacheFile.ModReference], *cacheFile)
	return cacheFile, nil
}

func readCacheFile(filename string) (*File, error) {
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	path := filepath.Join(downloadCache, filename)
	stat, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to stat file")
	}

	zipFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer zipFile.Close()

	size := stat.Size()
	reader, err := zip.NewReader(zipFile, size)
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

	hash, err := getFileHash(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get file hash")
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
			return nil, errors.Wrap(err, "failed to open icon file")
		}
		defer iconReader.Close()

		data, err := io.ReadAll(iconReader)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read icon file")
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
