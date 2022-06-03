package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/Khan/genqlient/graphql"
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
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

	copied := make(map[string][]string, len(constraints))
	for k, v := range constraints {
		copied[k] = []string{v}
	}

	instance := &resolvingInstance{
		Resolver:    d,
		InputLock:   lockFile,
		ToResolve:   copied,
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

	ToResolve map[string][]string

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
					Version:          constraint[0],
				})
			} else {
				if existingSML, ok := r.OutputLock[id]; ok {
					for _, cs := range constraint {
						smlVersionConstraint, _ := semver.NewConstraint(cs)
						if !smlVersionConstraint.Check(semver.MustParse(existingSML.Version)) {
							return errors.Errorf("mod %s version %s does not match constraint %s",
								id,
								existingSML.Version,
								constraint,
							)
						}
					}
				}

				var chosenSMLVersion *semver.Version
				for _, version := range r.SMLVersions.SmlVersions.Sml_versions {
					if version.Satisfactory_version > r.GameVersion {
						continue
					}

					currentVersion := semver.MustParse(version.Version)

					matches := true
					for _, cs := range constraint {
						smlVersionConstraint, _ := semver.NewConstraint(cs)

						if !smlVersionConstraint.Check(currentVersion) {
							matches = false
							break
						}
					}

					if matches {
						if chosenSMLVersion == nil || currentVersion.GreaterThan(chosenSMLVersion) {
							chosenSMLVersion = currentVersion
						}
					}
				}

				if chosenSMLVersion == nil {
					return errors.Errorf("could not find an SML version that matches constraint %s and game version %d", constraint, r.GameVersion)
				}

				r.OutputLock[id] = LockedMod{
					Version:      chosenSMLVersion.String(),
					Link:         fmt.Sprintf(smlDownloadTemplate, chosenSMLVersion.String()),
					Dependencies: map[string]string{},
				}
			}
		}

		nextResolve := make(map[string][]string)

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
			var selectedVersion ModVersion
			for _, version := range modVersions {
				matches := true

				for _, cs := range r.ToResolve[mod.Mod_reference] {
					resolvingConstraint, _ := semver.NewConstraint(cs)
					if !resolvingConstraint.Check(semver.MustParse(version.Version)) {
						matches = false
						break
					}
				}

				if matches {
					selectedVersion = version
					break
				}
			}

			if selectedVersion.Version == "" {
				return errors.Errorf("no version of %s matches constraints", mod.Mod_reference)
			}

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

				if resolving, ok := nextResolve[dependency.ModReference]; ok {
					constraint, _ := semver.NewConstraint(dependency.Constraint)

					for _, cs := range resolving {
						resolvingConstraint, _ := semver.NewConstraint(cs)
						intersects, _ := constraint.Intersects(resolvingConstraint)
						if !intersects {
							return errors.Errorf("mod %s constraint %s does not intersect with %s",
								dependency.ModReference,
								resolving,
								dependency.Constraint,
							)
						}
					}
				}

				if dependency.Optional {
					continue
				}

				nextResolve[dependency.ModReference] = append(nextResolve[dependency.ModReference], dependency.Constraint)
			}
		}

		r.ToResolve = nextResolve

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
		for modReference, constraints := range r.ToResolve {
			if _, ok := viewed[modReference]; ok {
				continue
			}

			viewed[modReference] = true

			if locked, ok := (*r.InputLock)[modReference]; ok {
				passes := true

				for _, cs := range constraints {
					constraint, _ := semver.NewConstraint(cs)
					if !constraint.Check(semver.MustParse(locked.Version)) {
						passes = false
						break
					}
				}

				if passes {
					delete(r.ToResolve, modReference)
					r.OutputLock[modReference] = locked
					for k, v := range locked.Dependencies {
						if alreadyResolving, ok := r.ToResolve[k]; ok {
							newConstraint, _ := semver.NewConstraint(v)
							for _, resolvingConstraint := range alreadyResolving {
								cs2, _ := semver.NewConstraint(resolvingConstraint)
								intersects, _ := newConstraint.Intersects(cs2)
								if !intersects {
									return errors.Errorf("mod %s constraint %s does not intersect with %s",
										k,
										v,
										alreadyResolving,
									)
								}
							}

							r.ToResolve[k] = append(r.ToResolve[k], v)

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

						r.ToResolve[k] = append(r.ToResolve[k], v)
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
