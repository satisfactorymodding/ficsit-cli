package cli

import (
	"github.com/Khan/genqlient/graphql"
	"github.com/pkg/errors"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type GlobalContext struct {
	Installations *Installations
	Profiles      *Profiles
	APIClient     graphql.Client
}

var globalContext *GlobalContext

func InitCLI(apiOnly bool) (*GlobalContext, error) {
	if globalContext != nil {
		return globalContext, nil
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

		globalContext = &GlobalContext{
			Installations: installations,
			Profiles:      profiles,
			APIClient:     ficsit.InitAPI(),
		}
	} else {
		globalContext = &GlobalContext{
			APIClient: ficsit.InitAPI(),
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
