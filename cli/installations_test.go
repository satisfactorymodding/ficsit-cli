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
	ctx, err := InitCLI()
	testza.AssertNoError(t, err)

	profileName := "InstallationTest"
	profile, err := ctx.Profiles.AddProfile(profileName)
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, profile.AddMod("AreaActions", ">=1.6.5"))
	testza.AssertNoError(t, profile.AddMod("ArmorModules__Modpack_All", ">=1.4.1"))

	serverLocation := os.Getenv("SF_DEDICATED_SERVER")
	if serverLocation != "" {
		installation, err := ctx.Installations.AddInstallation(ctx, serverLocation, profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx)
		testza.AssertNoError(t, err)
	}
}
