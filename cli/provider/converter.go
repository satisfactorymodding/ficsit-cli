package provider

import (
	resolver "github.com/satisfactorymodding/ficsit-resolver"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

func convertFicsitVersionsToResolver(versions []ficsit.ModVersion) []resolver.ModVersion {
	modVersions := make([]resolver.ModVersion, len(versions))
	for i, modVersion := range versions {
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
				TargetName: resolver.TargetName(target.TargetName),
				Link:       viper.GetString("api-base") + target.Link,
				Hash:       target.Hash,
				Size:       target.Size,
			}
		}

		modVersions[i] = resolver.ModVersion{
			Version:          modVersion.Version,
			GameVersion:      modVersion.GameVersion,
			Dependencies:     dependencies,
			Targets:          targets,
			RequiredOnRemote: modVersion.RequiredOnRemote,
		}
	}
	return modVersions
}
