package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/cli/cache"
	"github.com/satisfactorymodding/ficsit-cli/cli/localregistry"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type LocalProvider struct{}

func NewLocalProvider() LocalProvider {
	return LocalProvider{}
}

func (p LocalProvider) Mods(_ context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	cachedMods, err := cache.GetCacheMods()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	mods := make([]ficsit.ModsModsGetModsModsMod, 0)

	cachedMods.Range(func(modReference string, cachedMod cache.Mod) bool {
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
			Name:              cachedMod.Name,
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
	cachedMod, err := cache.GetCacheMod(modReference)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	authors := make([]ficsit.GetModModAuthorsUserMod, 0)

	for _, author := range strings.Split(cachedMod.Author, ",") {
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
			Name:             cachedMod.Name,
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
	modVersions, err := localregistry.GetModVersions(modID)
	if err != nil {
		return nil, fmt.Errorf("failed to get local mod versions: %w", err)
	}

	// TODO: only list as available the versions that have at least one target cached

	return convertFicsitVersionsToResolver(modVersions), nil
}

func (p LocalProvider) GetModName(_ context.Context, modReference string) (*resolver.ModName, error) {
	cachedMod, err := cache.GetCacheMod(modReference)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}

	return &resolver.ModName{
		ID:           modReference,
		Name:         cachedMod.Name,
		ModReference: modReference,
	}, nil
}

func (p LocalProvider) IsOffline() bool {
	return true
}
