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

func InitCLI() (*GlobalContext, error) {
	if globalContext != nil {
		return globalContext, nil
	}

	profiles, err := InitProfiles()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize profiles")
	}

	installations, err := InitInstallations()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize installations")
	}

	ctx := &GlobalContext{
		Installations: installations,
		Profiles:      profiles,
		APIClient:     ficsit.InitAPI(),
	}

	globalContext = ctx

	return ctx, nil
}

func (g *GlobalContext) Save() error {
	if err := g.Installations.Save(); err != nil {
		return err
	}

	if err := g.Profiles.Save(); err != nil {
		return err
	}

	return nil
}
