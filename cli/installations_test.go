package cli

import (
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/satisfactorymodding/ficsit-cli/cfg"
)

func init() {
	cfg.SetDefaults()
}

func TestInstallationsInit(t *testing.T) {
	installations, err := InitInstallations()
	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, installations)
}

func TestAddInstallation(t *testing.T) {
	ctx, err := InitCLI(false)
	testza.AssertNoError(t, err)

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	err = ctx.ReInit()
	testza.AssertNoError(t, err)

	ctx.Provider = MockProvider{}

	profileName := "InstallationTest"
	profile, err := ctx.Profiles.AddProfile(profileName)
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, profile.AddMod("AreaActions", ">=1.6.5"))
	testza.AssertNoError(t, profile.AddMod("RefinedPower", ">=3.2.10"))

	serverLocation := os.Getenv("SF_DEDICATED_SERVER")
	if serverLocation != "" {
		installation, err := ctx.Installations.AddInstallation(ctx, serverLocation, profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		installation.Vanilla = true
		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
	}
}
