package provider

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/satisfactorymodding/ficsit-cli/cli/cache"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type localProvider struct{}

func initLocalProvider() localProvider {
	return localProvider{}
}

func (p localProvider) Mods(_ context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	cachedMods, err := cache.GetCache()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache")
	}

	mods := make([]ficsit.ModsModsGetModsModsMod, 0)

	cachedMods.Range(func(modReference string, files []cache.File) bool {
		if modReference == "SML" {
			return true
		}

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

func (p localProvider) GetMod(_ context.Context, modReference string) (*ficsit.GetModResponse, error) {
	cachedModFiles, err := cache.GetCacheMod(modReference)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache")
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

func (p localProvider) ModVersions(_ context.Context, modReference string, filter ficsit.VersionFilter) (*ficsit.ModVersionsResponse, error) {
	cachedModFiles, err := cache.GetCacheMod(modReference)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache")
	}

	if filter.Limit == 0 {
		filter.Limit = 25
	}

	versions := make([]ficsit.ModVersionsModVersionsVersion, 0)

	for _, modFile := range cachedModFiles[filter.Offset : filter.Offset+filter.Limit] {
		versions = append(versions, ficsit.ModVersionsModVersionsVersion{
			Id:      modReference + ":" + modFile.Plugin.SemVersion,
			Version: modFile.Plugin.SemVersion,
		})
	}

	return &ficsit.ModVersionsResponse{
		Mod: ficsit.ModVersionsMod{
			Id:       modReference,
			Versions: versions,
		},
	}, nil
}

func (p localProvider) SMLVersions(_ context.Context) (*ficsit.SMLVersionsResponse, error) {
	cachedSMLFiles, err := cache.GetCacheMod("SML")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache")
	}

	smlVersions := make([]ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion, 0)

	for _, smlFile := range cachedSMLFiles {
		smlVersions = append(smlVersions, ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion{
			Id:                   "SML:" + smlFile.Plugin.SemVersion,
			Version:              smlFile.Plugin.SemVersion,
			Satisfactory_version: 0, // TODO: where can this be obtained from?
		})
	}

	return &ficsit.SMLVersionsResponse{
		SmlVersions: ficsit.SMLVersionsSmlVersionsGetSMLVersions{
			Count:        len(smlVersions),
			Sml_versions: smlVersions,
		},
	}, nil
}

func (p localProvider) ModVersionsWithDependencies(_ context.Context, modID string) (*ficsit.AllVersionsResponse, error) {
	cachedModFiles, err := cache.GetCacheMod(modID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache")
	}

	versions := make([]ficsit.ModVersion, 0)

	for _, modFile := range cachedModFiles {
		versions = append(versions, ficsit.ModVersion{
			ID:      modID + ":" + modFile.Plugin.SemVersion,
			Version: modFile.Plugin.SemVersion,
		})
	}

	return &ficsit.AllVersionsResponse{
		Success: true,
		Data:    versions,
	}, nil
}

func (p localProvider) GetModName(_ context.Context, modReference string) (*ficsit.GetModNameResponse, error) {
	cachedModFiles, err := cache.GetCacheMod(modReference)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache")
	}

	if len(cachedModFiles) == 0 {
		return nil, errors.New("mod not found")
	}

	return &ficsit.GetModNameResponse{
		Mod: ficsit.GetModNameMod{
			Id:            modReference,
			Name:          cachedModFiles[0].Plugin.FriendlyName,
			Mod_reference: modReference,
		},
	}, nil
}
