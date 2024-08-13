package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/cli/cache"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type LocalProvider struct{}

func NewLocalProvider() LocalProvider {
	return LocalProvider{}
}

func (p LocalProvider) Mods(_ context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	cachedMods, err := cache.GetCache()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	mods := make([]ficsit.ModsModsGetModsModsMod, 0)

	cachedMods.Range(func(modReference string, files []cache.File) bool {
		if len(filter.References) > 0 {
			skip := true

			for _, a := range filter.References {
				if a == modReference {
					skip = false
					break
				}
			}

			if skip {
				return true
			}
		}

		mods = append(mods, ficsit.ModsModsGetModsModsMod{
			Id:                modReference,
			Name:              files[0].Plugin.FriendlyName,
			Mod_reference:     modReference,
			Last_version_date: time.Now(),
			Created_at:        time.Now(),
			Downloads:         0,
			Popularity:        0,
			Hotness:           0,
		})

		return true
	})

	if filter.Limit == 0 {
		filter.Limit = 25
	}

	low := filter.Offset
	high := filter.Offset + filter.Limit

	if low > len(mods) {
		return &ficsit.ModsResponse{
			Mods: ficsit.ModsModsGetMods{
				Count: 0,
				Mods:  []ficsit.ModsModsGetModsModsMod{},
			},
		}, nil
	}

	if high > len(mods) {
		high = len(mods)
	}

	mods = mods[low:high]

	return &ficsit.ModsResponse{
		Mods: ficsit.ModsModsGetMods{
			Count: len(mods),
			Mods:  mods,
		},
	}, nil
}

func (p LocalProvider) GetMod(_ context.Context, modReference string) (*ficsit.GetModResponse, error) {
	cachedModFiles, err := cache.GetCacheMod(modReference)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	if len(cachedModFiles) == 0 {
		return nil, errors.New("mod not found")
	}

	authors := make([]ficsit.GetModModAuthorsUserMod, 0)

	for _, author := range strings.Split(cachedModFiles[0].Plugin.CreatedBy, ",") {
		authors = append(authors, ficsit.GetModModAuthorsUserMod{
			Role: "Unknown",
			User: ficsit.GetModModAuthorsUserModUser{
				Username: author,
			},
		})
	}

	return &ficsit.GetModResponse{
		Mod: ficsit.GetModMod{
			Id:               modReference,
			Name:             cachedModFiles[0].Plugin.FriendlyName,
			Mod_reference:    modReference,
			Created_at:       time.Now(),
			Views:            0,
			Downloads:        0,
			Authors:          authors,
			Full_description: "",
			Source_url:       "",
		},
	}, nil
}

func (p LocalProvider) ModVersionsWithDependencies(_ context.Context, modID string) ([]resolver.ModVersion, error) {
	cachedModFiles, err := cache.GetCacheMod(modID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	versions := make([]resolver.ModVersion, 0)

	for _, modFile := range cachedModFiles {
		versions = append(versions, resolver.ModVersion{
			ID:          modID + ":" + modFile.Plugin.SemVersion,
			Version:     modFile.Plugin.SemVersion,
			GameVersion: modFile.Plugin.GameVersion,
		})
	}

	return versions, nil
}

func (p LocalProvider) GetModName(_ context.Context, modReference string) (*resolver.ModName, error) {
	cachedModFiles, err := cache.GetCacheMod(modReference)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	if len(cachedModFiles) == 0 {
		return nil, errors.New("mod not found")
	}

	return &resolver.ModName{
		ID:           modReference,
		Name:         cachedModFiles[0].Plugin.FriendlyName,
		ModReference: modReference,
	}, nil
}

func (p LocalProvider) IsOffline() bool {
	return true
}
