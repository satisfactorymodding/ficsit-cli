package cli

import "github.com/pkg/errors"

type GlobalContext struct {
	Installations *Installations
	Profiles      *Profiles
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
