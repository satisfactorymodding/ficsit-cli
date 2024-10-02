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

	"github.com/mircearoata/pubgrub-go/pubgrub/semver"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/viper"
)

const IconFilename = "Resources/Icon128.png" // This is the path UE expects for the icon

type Mod struct {
	ModReference  string
	Name          string
	Author        string
	Icon          *string
	LatestVersion string
}

var loadedMods *xsync.MapOf[string, Mod]

func GetCacheMods() (*xsync.MapOf[string, Mod], error) {
	if loadedMods != nil {
		return loadedMods, nil
	}
	return LoadCacheMods()
}

func GetCacheMod(mod string) (Mod, error) {
	cache, err := GetCacheMods()
	if err != nil {
		return Mod{}, err
	}
	value, _ := cache.Load(mod)
	return value, nil
}

func LoadCacheMods() (*xsync.MapOf[string, Mod], error) {
	loadedMods = xsync.NewMapOf[string, Mod]()
	downloadCache := filepath.Join(viper.GetString("cache-dir"), "downloadCache")
	if _, err := os.Stat(downloadCache); os.IsNotExist(err) {
		return loadedMods, nil
	}

	items, err := os.ReadDir(downloadCache)
	if err != nil {
		return nil, fmt.Errorf("failed reading download cache: %w", err)
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}

		_, err = addFileToCache(item.Name())
		if err != nil {
			slog.Error("failed to add file to cache", slog.String("file", item.Name()), slog.Any("err", err))
		}
	}
	return loadedMods, nil
}

func addFileToCache(filename string) (*Mod, error) {
	cacheFile, err := readCacheFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	loadedMods.Compute(cacheFile.ModReference, func(oldValue Mod, loaded bool) (Mod, bool) {
		if !loaded {
			return *cacheFile, false
		}
		oldVersion, err := semver.NewVersion(oldValue.LatestVersion)
		if err != nil {
			slog.Error("failed to parse version", slog.String("version", oldValue.LatestVersion), slog.Any("err", err))
			return *cacheFile, false
		}
		newVersion, err := semver.NewVersion(cacheFile.LatestVersion)
		if err != nil {
			slog.Error("failed to parse version", slog.String("version", cacheFile.LatestVersion), slog.Any("err", err))
			return oldValue, false
		}
		if newVersion.Compare(oldVersion) > 0 {
			return *cacheFile, false
		}
		return oldValue, false
	})

	return cacheFile, nil
}

func readCacheFile(filename string) (*Mod, error) {
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

	return &Mod{
		ModReference:  modReference,
		Name:          uplugin.FriendlyName,
		Author:        uplugin.CreatedBy,
		Icon:          icon,
		LatestVersion: uplugin.SemVersion,
	}, nil
}
