package provider

import (
	"context"
	"errors"

	"github.com/Khan/genqlient/graphql"
	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type FicsitProvider struct {
	client graphql.Client
}

func NewFicsitProvider(client graphql.Client) FicsitProvider {
	return FicsitProvider{
		client,
	}
}

func (p FicsitProvider) Mods(context context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	return ficsit.Mods(context, p.client, filter)
}

func (p FicsitProvider) GetMod(context context.Context, modReference string) (*ficsit.GetModResponse, error) {
	return ficsit.GetMod(context, p.client, modReference)
}

func (p FicsitProvider) ModVersions(context context.Context, modReference string, filter ficsit.VersionFilter) (*ficsit.ModVersionsResponse, error) {
	return ficsit.ModVersions(context, p.client, modReference, filter)
}

func (p FicsitProvider) SMLVersions(context context.Context) ([]resolver.SMLVersion, error) {
	response, err := ficsit.SMLVersions(context, p.client)
	if err != nil {
		return nil, err
	}

	smlVersions := make([]resolver.SMLVersion, len(response.SmlVersions.Sml_versions))
	for i, version := range response.GetSmlVersions().Sml_versions {
		targets := make([]resolver.SMLVersionTarget, len(version.Targets))

		for j, target := range version.Targets {
			targets[j] = resolver.SMLVersionTarget{
				TargetName: resolver.TargetName(target.TargetName),
				Link:       target.Link,
			}
		}

		smlVersions[i] = resolver.SMLVersion{
			ID:                  version.Id,
			Version:             version.Version,
			SatisfactoryVersion: version.Satisfactory_version,
			Targets:             targets,
		}
	}

	return smlVersions, nil
}

func (p FicsitProvider) ModVersionsWithDependencies(_ context.Context, modID string) ([]resolver.ModVersion, error) {
	response, err := ficsit.GetAllModVersions(modID)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, errors.New(response.Error.Message)
	}

	modVersions := make([]resolver.ModVersion, len(response.Data))
	for i, modVersion := range response.Data {
		dependencies := make([]resolver.Dependency, len(modVersion.Dependencies))
		for j, dependency := range modVersion.Dependencies {
			dependencies[j] = resolver.Dependency{
				ModID:     dependency.ModID,
				Condition: dependency.Condition,
				Optional:  dependency.Optional,
			}
		}

		targets := make([]resolver.Target, len(modVersion.Targets))
		for j, target := range modVersion.Targets {
			targets[j] = resolver.Target{
				VersionID:  target.VersionID,
				TargetName: resolver.TargetName(target.TargetName),
				Hash:       target.Hash,
				Size:       target.Size,
			}
		}

		modVersions[i] = resolver.ModVersion{
			ID:           modVersion.ID,
			Version:      modVersion.Version,
			Dependencies: dependencies,
			Targets:      targets,
		}
	}

	return modVersions, err
}

func (p FicsitProvider) GetModName(context context.Context, modReference string) (*resolver.ModName, error) {
	response, err := ficsit.GetModName(context, p.client, modReference)
	if err != nil {
		return nil, err
	}

	return &resolver.ModName{
		ID:           response.Mod.Id,
		ModReference: response.Mod.Mod_reference,
		Name:         response.Mod.Name,
	}, nil
}

func (p FicsitProvider) IsOffline() bool {
	return false
}
