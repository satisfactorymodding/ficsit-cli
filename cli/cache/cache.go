package cache

import (
	"archive/zip"
	"encoding/json"
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

func GetCache() (map[string][]CacheFile, error) {
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if _, err := os.Stat(downloadCache); os.IsNotExist(err) {
		return map[string][]CacheFile{}, nil
	}

	items, err := os.ReadDir(downloadCache)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading download cache")
	}

	mods := map[string][]CacheFile{}

	for _, item := range items {
		if item.IsDir() {
			continue
		}

		fullpath := filepath.Join(downloadCache, item.Name())

		stat, err := os.Stat(fullpath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to stat file: "+fullpath)
		}

		zipFile, err := os.Open(fullpath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file: "+fullpath)
		}

		reader, err := zip.NewReader(zipFile, stat.Size())
		if err != nil {
			return nil, errors.Wrap(err, "failed to read zip: "+fullpath)
		}

		var upluginFile *zip.File
		for _, file := range reader.File {
			if strings.HasSuffix(file.Name, ".uplugin") {
				upluginFile = file
				break
			}
		}
		if upluginFile == nil {
			return nil, errors.New("no uplugin file found in zip: " + fullpath)
		}

		f, err := upluginFile.Open()
		if err != nil {
			return nil, errors.Wrap(err, "failed to open uplugin file: "+fullpath)
		}

		var uplugin UPlugin
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read uplugin file: "+fullpath)
		}
		if err := json.Unmarshal(data, &uplugin); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal uplugin file: "+fullpath)
		}

		modReference := strings.TrimSuffix(upluginFile.Name, ".uplugin")

		hash, err := utils.SHA256Data(f)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash uplugin file: "+fullpath)
		}

		if _, ok := mods[modReference]; !ok {
			mods[modReference] = []CacheFile{}
		}
		mods[modReference] = append(mods[modReference], CacheFile{
			ModReference: modReference,
			Hash:         hash,
			Plugin:       uplugin,
		})
	}
	return mods, nil
}

func GetCacheMod(mod string) ([]CacheFile, error) {
	cache, err := GetCache()
	if err != nil {
		return nil, err
	}
	return cache[mod], nil
}
