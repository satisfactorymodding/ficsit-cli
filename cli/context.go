package cli

import (
	"fmt"
	"log/slog"

	"github.com/Khan/genqlient/graphql"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli/cache"
	"github.com/satisfactorymodding/ficsit-cli/cli/localregistry"
	"github.com/satisfactorymodding/ficsit-cli/cli/provider"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type GlobalContext struct {
	Installations *Installations
	Profiles      *Profiles
	APIClient     graphql.Client
	Provider      provider.Provider
}

var globalContext *GlobalContext

func InitCLI(apiOnly bool) (*GlobalContext, error) {
	if globalContext != nil {
		return globalContext, nil
	}

	apiClient := ficsit.InitAPI()

	mixedProvider := provider.InitMixedProvider(provider.NewFicsitProvider(apiClient), provider.NewLocalProvider())

	if viper.GetBool("offline") {
		mixedProvider.Offline = true
	}

	if !apiOnly {
		profiles, err := InitProfiles()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize profiles: %w", err)
		}

		installations, err := InitInstallations()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize installations: %w", err)
		}

		_, err = cache.LoadCacheMods()
		if err != nil {
			return nil, fmt.Errorf("failed to load cache: %w", err)
		}

		err = localregistry.Init()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize local registry: %w", err)
		}

		globalContext = &GlobalContext{
			Installations: installations,
			Profiles:      profiles,
			APIClient:     apiClient,
			Provider:      mixedProvider,
		}
	} else {
		globalContext = &GlobalContext{
			APIClient: apiClient,
			Provider:  mixedProvider,
		}
	}

	return globalContext, nil
}

// ReInit will initialize the context
//
// Used only by tests
func (g *GlobalContext) ReInit() error {
	profiles, err := InitProfiles()
	if err != nil {
		return fmt.Errorf("failed to initialize profiles: %w", err)
	}

	installations, err := InitInstallations()
	if err != nil {
		return fmt.Errorf("failed to initialize installations: %w", err)
	}

	g.Installations = installations
	g.Profiles = profiles

	return g.Save()
}

// Wipe will remove any trace of ficsit anywhere
func (g *GlobalContext) Wipe() error {
	slog.Info("wiping global context")

	// Wipe all installations
	for _, installation := range g.Installations.Installations {
		if err := installation.Wipe(); err != nil {
			return fmt.Errorf("failed wiping installation: %w", err)
		}

		if err := g.Installations.DeleteInstallation(installation.Path); err != nil {
			return fmt.Errorf("failed deleting installation: %w", err)
		}
	}

	// Wipe all profiles
	for _, profile := range g.Profiles.Profiles {
		if err := g.Profiles.DeleteProfile(profile.Name); err != nil {
			return fmt.Errorf("failed deleting profile: %w", err)
		}
	}

	return g.Save()
}

func (g *GlobalContext) Save() error {
	if err := g.Installations.Save(); err != nil {
		return fmt.Errorf("failed to save installations: %w", err)
	}

	if err := g.Profiles.Save(); err != nil {
		return fmt.Errorf("failed to save profiles: %w", err)
	}

	return nil
}
