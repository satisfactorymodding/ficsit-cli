package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/Khan/genqlient/graphql"
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/utils"
	"github.com/spf13/viper"
)

const smlDownloadTemplate = `https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/%s/SML.zip`

type DependencyResolver struct {
	apiClient graphql.Client
}

func NewDependencyResolver(apiClient graphql.Client) DependencyResolver {
	return DependencyResolver{apiClient: apiClient}
}

func (d DependencyResolver) ResolveModDependencies(constraints map[string]string, lockFile *LockFile, gameVersion int) (LockFile, error) {
	smlVersionsDB, err := ficsit.SMLVersions(context.TODO(), d.apiClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed fetching SMl versions")
	}

	instance := &resolvingInstance{
		Resolver:    d,
		InputLock:   lockFile,
		ToResolve:   utils.CopyMap(constraints),
		OutputLock:  make(LockFile),
		SMLVersions: smlVersionsDB,
		GameVersion: gameVersion,
	}

	if err := instance.Step(); err != nil {
		return nil, err
	}

	return instance.OutputLock, nil
}

type resolvingInstance struct {
	Resolver DependencyResolver

	InputLock *LockFile

	ToResolve map[string]string

	OutputLock LockFile

	SMLVersions *ficsit.SMLVersionsResponse
	GameVersion int
}

func (r *resolvingInstance) Step() error {
	if len(r.ToResolve) > 0 {
		if err := r.LockStep(make(map[string]bool)); err != nil {
			return err
		}

		converted := make([]ficsit.ModVersionConstraint, 0)
		for id, constraint := range r.ToResolve {
			if id != "SML" {
				converted = append(converted, ficsit.ModVersionConstraint{
					ModIdOrReference: id,
					Version:          constraint,
				})
			} else {
				smlVersionConstraint, _ := semver.NewConstraint(constraint)
				if existingSML, ok := r.OutputLock[id]; ok {
					if !smlVersionConstraint.Check(semver.MustParse(existingSML.Version)) {
						return errors.New("failed resolving dependencies. requires different versions of " + id)
					}
				}

				var chosenSMLVersion *semver.Version
				for _, version := range r.SMLVersions.SmlVersions.Sml_versions {
					if version.Satisfactory_version > r.GameVersion {
						continue
					}

					currentVersion := semver.MustParse(version.Version)
					if smlVersionConstraint.Check(currentVersion) {
						if chosenSMLVersion == nil || currentVersion.GreaterThan(chosenSMLVersion) {
							chosenSMLVersion = currentVersion
						}
					}
				}

				if chosenSMLVersion == nil {
					return fmt.Errorf("could not find an SML version that matches constraint %s and game version %d", constraint, r.GameVersion)
				}

				r.OutputLock[id] = LockedMod{
					Version:      chosenSMLVersion.String(),
					Link:         fmt.Sprintf(smlDownloadTemplate, chosenSMLVersion.String()),
					Dependencies: map[string]string{},
				}
			}
		}

		r.ToResolve = make(map[string]string)

		// TODO Cache
		dependencies, err := ficsit.ResolveModDependencies(context.TODO(), r.Resolver.apiClient, converted)
		if err != nil {
			return errors.Wrap(err, "failed resolving mod dependencies")
		}

		for _, mod := range dependencies.Mods {
			modVersions := make([]ModVersion, len(mod.Versions))
			for i, version := range mod.Versions {
				versionDependencies := make([]VersionDependency, len(version.Dependencies))

				for j, dependency := range version.Dependencies {
					versionDependencies[j] = VersionDependency{
						ModReference: dependency.Mod_id,
						Constraint:   dependency.Condition,
						Optional:     dependency.Optional,
					}
				}

				modVersions[i] = ModVersion{
					ID:           version.Id,
					Version:      version.Version,
					Link:         viper.GetString("api-base") + version.Link,
					Hash:         version.Hash,
					Dependencies: versionDependencies,
				}
			}

			sort.Slice(modVersions, func(i, j int) bool {
				a := semver.MustParse(modVersions[i].Version)
				b := semver.MustParse(modVersions[j].Version)
				return b.LessThan(a)
			})

			// Pick latest version
			// TODO Clone and branch
			selectedVersion := modVersions[0]

			if _, ok := r.OutputLock[mod.Mod_reference]; ok {
				if r.OutputLock[mod.Mod_reference].Version != selectedVersion.Version {
					return errors.New("failed resolving dependencies. requires different versions of " + mod.Mod_reference)
				}
			}

			modDependencies := make(map[string]string)
			for _, dependency := range selectedVersion.Dependencies {
				if !dependency.Optional {
					modDependencies[dependency.ModReference] = dependency.Constraint
				}
			}

			r.OutputLock[mod.Mod_reference] = LockedMod{
				Version:      selectedVersion.Version,
				Hash:         selectedVersion.Hash,
				Link:         selectedVersion.Link,
				Dependencies: modDependencies,
			}

			for _, dependency := range selectedVersion.Dependencies {
				if previousSelectedVersion, ok := r.OutputLock[dependency.ModReference]; ok {
					constraint, _ := semver.NewConstraint(dependency.Constraint)
					if !constraint.Check(semver.MustParse(previousSelectedVersion.Version)) {
						return errors.Errorf("mod %s version %s does not match constraint %s",
							dependency.ModReference,
							previousSelectedVersion.Version,
							dependency.Constraint,
						)
					}
				}

				if resolving, ok := r.ToResolve[dependency.ModReference]; ok {
					constraint, _ := semver.NewConstraint(dependency.Constraint)
					resolvingConstraint, _ := semver.NewConstraint(resolving)
					intersects, _ := constraint.Intersects(resolvingConstraint)
					if !intersects {
						return errors.Errorf("mod %s constraint %s does not intersect with %s",
							dependency.ModReference,
							resolving,
							dependency.Constraint,
						)
					}
				}

				if dependency.Optional {
					continue
				}

				r.ToResolve[dependency.ModReference] = dependency.Constraint
			}
		}

		for _, constraint := range converted {
			if _, ok := r.OutputLock[constraint.ModIdOrReference]; !ok {
				return errors.New("failed resolving dependency: " + constraint.ModIdOrReference)
			}
		}
	}

	if len(r.ToResolve) > 0 {
		if err := r.Step(); err != nil {
			return err
		}
	}

	return nil
}

func (r *resolvingInstance) LockStep(viewed map[string]bool) error {
	added := false
	if r.InputLock != nil {
		for modReference, version := range r.ToResolve {
			if _, ok := viewed[modReference]; ok {
				continue
			}

			viewed[modReference] = true

			if locked, ok := (*r.InputLock)[modReference]; ok {
				constraint, _ := semver.NewConstraint(version)
				if constraint.Check(semver.MustParse(locked.Version)) {
					delete(r.ToResolve, modReference)
					r.OutputLock[modReference] = locked
					for k, v := range locked.Dependencies {
						if alreadyResolving, ok := r.ToResolve[k]; ok {
							cs1, _ := semver.NewConstraint(v)
							cs2, _ := semver.NewConstraint(alreadyResolving)
							intersects, _ := cs1.Intersects(cs2)
							if !intersects {
								return errors.Errorf("mod %s constraint %s does not intersect with %s",
									k,
									v,
									alreadyResolving,
								)
							}
							continue
						}

						if outVersion, ok := r.OutputLock[k]; ok {
							constraint, _ := semver.NewConstraint(v)
							if !constraint.Check(semver.MustParse(outVersion.Version)) {
								return errors.Errorf("mod %s version %s does not match constraint %s",
									k,
									outVersion.Version,
									v,
								)
							}
							continue
						}

						r.ToResolve[k] = v
						added = true
					}
				}
			}
		}
	}
	if added {
		if err := r.LockStep(viewed); err != nil {
			return err
		}
	}
	return nil
}
