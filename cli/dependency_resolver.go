package cli

import (
	"context"
	"fmt"
	"slices"

	"github.com/mircearoata/pubgrub-go/pubgrub"
	"github.com/mircearoata/pubgrub-go/pubgrub/helpers"
	"github.com/mircearoata/pubgrub-go/pubgrub/semver"
	"github.com/pkg/errors"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli/provider"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

const (
	rootPkg        = "$$root$$"
	smlPkg         = "SML"
	factoryGamePkg = "FactoryGame"
)

type DependencyResolver struct {
	provider provider.Provider
}

func NewDependencyResolver(provider provider.Provider) DependencyResolver {
	return DependencyResolver{provider}
}

type ficsitAPISource struct {
	provider       provider.Provider
	lockfile       *LockFile
	toInstall      map[string]semver.Constraint
	modVersionInfo *xsync.MapOf[string, ficsit.ModVersionsWithDependenciesResponse]
	gameVersion    semver.Version
	smlVersions    []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion
}

func (f *ficsitAPISource) GetPackageVersions(pkg string) ([]pubgrub.PackageVersion, error) {
	if pkg == rootPkg {
		return []pubgrub.PackageVersion{{Version: semver.Version{}, Dependencies: f.toInstall}}, nil
	}
	if pkg == factoryGamePkg {
		return []pubgrub.PackageVersion{{Version: f.gameVersion}}, nil
	}
	if pkg == smlPkg {
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
					factoryGamePkg: gameConstraint,
				},
			}
		}
		return versions, nil
	}
	response, err := f.provider.ModVersionsWithDependencies(context.TODO(), pkg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch mod %s", pkg)
	}
	if response.Mod.Id == "" {
		return nil, errors.Errorf("mod %s not found", pkg)
	}
	f.modVersionInfo.Store(pkg, *response)
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

func (f *ficsitAPISource) PickVersion(pkg string, versions []semver.Version) semver.Version {
	if f.lockfile != nil {
		if existing, ok := f.lockfile.Mods[pkg]; ok {
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

func (d DependencyResolver) ResolveModDependencies(constraints map[string]string, lockFile *LockFile, gameVersion int) (*LockFile, error) {
	smlVersionsDB, err := d.provider.SMLVersions(context.TODO())
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

	ficsitSource := &ficsitAPISource{
		provider:       d.provider,
		smlVersions:    smlVersionsDB.SmlVersions.Sml_versions,
		gameVersion:    gameVersionSemver,
		lockfile:       lockFile,
		toInstall:      toInstall,
		modVersionInfo: xsync.NewMapOf[string, ficsit.ModVersionsWithDependenciesResponse](),
	}

	result, err := pubgrub.Solve(helpers.NewCachingSource(ficsitSource), rootPkg)
	if err != nil {
		finalError := err
		var solverErr pubgrub.SolvingError
		if errors.As(err, &solverErr) {
			finalError = DependencyResolverError{SolvingError: solverErr, provider: d.provider, smlVersions: smlVersionsDB.SmlVersions.Sml_versions, gameVersion: gameVersion}
		}
		return nil, errors.Wrap(finalError, "failed to solve dependencies")
	}
	delete(result, rootPkg)
	delete(result, factoryGamePkg)

	outputLock := MakeLockfile()
	for k, v := range result {
		if k == smlPkg {
			for _, version := range ficsitSource.smlVersions {
				if version.Version == v.String() {
					targets := make(map[string]LockedModTarget)
					for _, target := range version.Targets {
						targets[string(target.TargetName)] = LockedModTarget{
							Link: target.Link,
						}
					}

					outputLock.Mods[k] = LockedMod{
						Version: v.String(),
						Targets: targets,
					}
					break
				}
			}
			continue
		}

		value, _ := ficsitSource.modVersionInfo.Load(k)
		versions := value.Mod.Versions
		for _, ver := range versions {
			if ver.Version == v.RawString() {
				targets := make(map[string]LockedModTarget)
				for _, target := range ver.Targets {
					targets[string(target.TargetName)] = LockedModTarget{
						Link: viper.GetString("api-base") + target.Link,
						Hash: target.Hash,
					}
				}

				outputLock.Mods[k] = LockedMod{
					Version: v.String(),
					Targets: targets,
				}
				break
			}
		}
	}

	return outputLock, nil
}
