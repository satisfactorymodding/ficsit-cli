package cli

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli/cache"
	"github.com/satisfactorymodding/ficsit-cli/cli/provider"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type GlobalContext struct {
	Installations *Installations
	Profiles      *Profiles
	APIClient     graphql.Client
	Provider      *provider.MixedProvider
}

var globalContext *GlobalContext

func InitCLI(apiOnly bool) (*GlobalContext, error) {
	if globalContext != nil {
		return globalContext, nil
	}

	apiClient := ficsit.InitAPI()

	mixedProvider := provider.InitMixedProvider(apiClient)

	if viper.GetBool("offline") {
		mixedProvider.Offline = true
	}

	if !apiOnly {
		profiles, err := InitProfiles()
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize profiles")
		}

		installations, err := InitInstallations()
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize installations")
		}

		_, err = cache.LoadCache()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load cache")
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

func (g *GlobalContext) Save() error {
	if err := g.Installations.Save(); err != nil {
		return errors.Wrap(err, "failed to save installations")
	}

	if err := g.Profiles.Save(); err != nil {
		return errors.Wrap(err, "failed to save profiles")
	}

	return nil
}
