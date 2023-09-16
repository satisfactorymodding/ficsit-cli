package cli

import (
	"context"
	"fmt"
	"slices"

	"github.com/Khan/genqlient/graphql"
	"github.com/mircearoata/pubgrub-go/pubgrub"
	"github.com/mircearoata/pubgrub-go/pubgrub/helpers"
	"github.com/mircearoata/pubgrub-go/pubgrub/semver"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

const smlDownloadTemplate = `https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v%s/SML.zip`

type DependencyResolver struct {
	apiClient graphql.Client
}

func NewDependencyResolver(apiClient graphql.Client) DependencyResolver {
	return DependencyResolver{apiClient: apiClient}
}

var rootPkg = "$$root$$"

type ficsitApiSource struct {
	apiClient   graphql.Client
	smlVersions []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion
	gameVersion semver.Version
	lockfile    *LockFile
	toInstall   map[string]semver.Constraint
}

func (f ficsitApiSource) GetPackageVersions(pkg string) ([]pubgrub.PackageVersion, error) {
	if pkg == rootPkg {
		return []pubgrub.PackageVersion{{Version: semver.Version{}, Dependencies: f.toInstall}}, nil
	}
	if pkg == "FactoryGame" {
		return []pubgrub.PackageVersion{{Version: f.gameVersion}}, nil
	}
	if pkg == "SML" {
		versions := make([]pubgrub.PackageVersion, len(f.smlVersions))
		for i, smlVersion := range f.smlVersions {
			v, err := semver.NewVersion(smlVersion.Version)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse version %s", smlVersion.Version)
			}
			gameConstraint, err := semver.NewConstraint(fmt.Sprintf(">=%d", smlVersion.Satisfactory_version))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse constraint %s", fmt.Sprintf(">=%d", smlVersion.Satisfactory_version))
			}
			versions[i] = pubgrub.PackageVersion{
				Version: v,
				Dependencies: map[string]semver.Constraint{
					"FactoryGame": gameConstraint,
				},
			}
		}
		return versions, nil
	}
	response, err := ficsit.ModVersionsWithDependencies(context.TODO(), f.apiClient, pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch mod %s", pkg)
	}
	if response.Mod.Id == "" {
		return nil, errors.Errorf("mod %s not found", pkg)
	}
	versions := make([]pubgrub.PackageVersion, len(response.Mod.Versions))
	for i, modVersion := range response.Mod.Versions {
		v, err := semver.NewVersion(modVersion.Version)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse version %s", modVersion.Version)
		}
		dependencies := make(map[string]semver.Constraint)
		optionalDependencies := make(map[string]semver.Constraint)
		for _, dependency := range modVersion.Dependencies {
			c, err := semver.NewConstraint(dependency.Condition)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse constraint %s", dependency.Condition)
			}
			if dependency.Optional {
				optionalDependencies[dependency.Mod_id] = c
			} else {
				dependencies[dependency.Mod_id] = c
			}
		}
		versions[i] = pubgrub.PackageVersion{
			Version:              v,
			Dependencies:         dependencies,
			OptionalDependencies: optionalDependencies,
		}
	}
	return versions, nil
}

func (f ficsitApiSource) PickVersion(pkg string, versions []semver.Version) semver.Version {
	if f.lockfile != nil {
		if existing, ok := (*f.lockfile)[pkg]; ok {
			v, err := semver.NewVersion(existing.Version)
			if err == nil {
				if slices.ContainsFunc(versions, func(version semver.Version) bool {
					return v.Compare(version) == 0
				}) {
					return v
				}
			}
		}
	}
	return helpers.StandardVersionPriority(versions)
}

func (d DependencyResolver) ResolveModDependencies(constraints map[string]string, lockFile *LockFile, gameVersion int) (LockFile, error) {
	smlVersionsDB, err := ficsit.SMLVersions(context.TODO(), d.apiClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed fetching SML versions")
	}

	gameVersionSemver, err := semver.NewVersion(fmt.Sprintf("%d", gameVersion))
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing game version")
	}

	toInstall := make(map[string]semver.Constraint, len(constraints))
	for k, v := range constraints {
		c, err := semver.NewConstraint(v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse constraint %s", v)
		}
		toInstall[k] = c
	}

	source := helpers.NewCachingSource(ficsitApiSource{
		apiClient:   d.apiClient,
		smlVersions: smlVersionsDB.SmlVersions.Sml_versions,
		gameVersion: gameVersionSemver,
		lockfile:    lockFile,
		toInstall:   toInstall,
	})

	result, err := pubgrub.Solve(source, rootPkg)
	if err != nil {
		finalError := err
		var solverErr pubgrub.SolvingError
		if errors.As(err, &solverErr) {
			finalError = DependencyResolverError{SolvingError: solverErr, apiClient: d.apiClient, smlVersions: smlVersionsDB.SmlVersions.Sml_versions, gameVersion: gameVersion}
		}
		return nil, errors.Wrap(finalError, "failed to solve dependencies")
	}
	delete(result, rootPkg)
	delete(result, "FactoryGame")

	outputLock := make(LockFile, len(result))
	for k, v := range result {
		if k == "SML" {
			outputLock[k] = LockedMod{
				Version: v.String(),
				Hash:    "",
				Link:    fmt.Sprintf(smlDownloadTemplate, v.String()),
			}
			continue
		}
		versionResponse, err := ficsit.Version(context.TODO(), d.apiClient, k, v.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch version")
		}

		outputLock[k] = LockedMod{
			Version: v.String(),
			Hash:    versionResponse.Mod.Version.Hash,
			Link:    viper.GetString("api-base") + versionResponse.Mod.Version.Link,
		}
	}

	return outputLock, nil
}
