package cli

import (
	"context"
	"sort"

	"github.com/Khan/genqlient/graphql"
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type DependencyResolver struct {
	apiClient graphql.Client
}

func NewDependencyResolver(apiClient graphql.Client) DependencyResolver {
	return DependencyResolver{apiClient: apiClient}
}

func (d DependencyResolver) ResolveModDependencies(constraints map[string]string) (map[string]ModVersion, error) {
	results := make(map[string]ModVersion)

	toResolve := constraints

	for len(toResolve) > 0 {
		converted := make([]ficsit.ModVersionConstraint, 0)
		for id, constraint := range toResolve {
			converted = append(converted, ficsit.ModVersionConstraint{
				ModIdOrReference: id,
				Version:          constraint,
			})
		}

		toResolve = make(map[string]string)

		dependencies, err := ficsit.ResolveModDependencies(context.TODO(), d.apiClient, converted)
		if err != nil {
			return nil, errors.Wrap(err, "failed resolving mod dependencies")
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
					Dependencies: versionDependencies,
				}
			}

			sort.Slice(modVersions, func(i, j int) bool {
				a := semver.MustParse(modVersions[i].Version)
				b := semver.MustParse(modVersions[j].Version)
				return b.LessThan(a)
			})

			// Pick latest version
			selectedVersion := modVersions[0]

			if _, ok := results[mod.Mod_reference]; ok {
				if results[mod.Mod_reference].Version != selectedVersion.Version {
					return nil, errors.New("failed resolving dependencies. requires different versions of " + mod.Mod_reference)
				}
			}

			results[mod.Mod_reference] = selectedVersion

			for _, dependency := range selectedVersion.Dependencies {
				if previousSelectedVersion, ok := results[dependency.ModReference]; ok {
					constraint, _ := semver.NewConstraint(dependency.Constraint)
					if !constraint.Check(semver.MustParse(previousSelectedVersion.Version)) {
						return nil, errors.Errorf("mod %s version %s does not match constraint %s",
							dependency.ModReference,
							previousSelectedVersion.Version,
							dependency.Constraint,
						)
					}
				}

				// TODO If already exists, verify which constraint is newer and use that
				toResolve[dependency.ModReference] = dependency.Constraint
			}
		}

		for _, constraint := range converted {
			// Ignore SML
			if constraint.ModIdOrReference == "SML" {
				continue
			}

			if _, ok := results[constraint.ModIdOrReference]; !ok {
				return nil, errors.New("failed resolving dependency: " + constraint.ModIdOrReference)
			}
		}
	}

	return results, nil
}
