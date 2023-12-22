package cli

import (
	"github.com/davecgh/go-spew/spew"
	"os"
	"path/filepath"
	"runtime"
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

func TestAddLocalInstallation(t *testing.T) {
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
	testza.AssertNoError(t, profile.AddMod("AreaActions", "1.6.5"))
	testza.AssertNoError(t, profile.AddMod("RefinedPower", "3.2.10"))

	serverLocation := os.Getenv("SF_DEDICATED_SERVER")
	if serverLocation != "" {
		testza.AssertNoError(t, os.RemoveAll(filepath.Join(serverLocation, "FactoryGame", "Mods")))

		dir, err := os.ReadDir(filepath.Join(serverLocation, "FactoryGame", "Mods"))
		testza.AssertNoError(t, err)
		spew.Dump(dir)

		installation, err := ctx.Installations.AddInstallation(ctx, serverLocation, profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		installation.Vanilla = true
		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
	}

	err = ctx.Wipe()
	testza.AssertNoError(t, err)
}

func TestAddFTPInstallation(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Not supported
		return
	}

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
	testza.AssertNoError(t, profile.AddMod("AreaActions", "1.6.5"))
	testza.AssertNoError(t, profile.AddMod("RefinedPower", "3.2.10"))

	serverLocation := os.Getenv("SF_DEDICATED_SERVER")
	if serverLocation != "" {
		testza.AssertNoError(t, os.RemoveAll(filepath.Join(serverLocation, "FactoryGame", "Mods")))

		dir, err := os.ReadDir(filepath.Join(serverLocation, "FactoryGame", "Mods"))
		testza.AssertNoError(t, err)
		spew.Dump(dir)

		installation, err := ctx.Installations.AddInstallation(ctx, "ftp://user:pass@localhost:2121/server", profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		dir, err = os.ReadDir(filepath.Join(serverLocation, "FactoryGame", "Mods"))
		testza.AssertNoError(t, err)
		spew.Dump(dir)

		installation.Vanilla = true
		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
	}

	err = ctx.Wipe()
	testza.AssertNoError(t, err)
}

func TestAddSFTPInstallation(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Not supported
		return
	}

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
	testza.AssertNoError(t, profile.AddMod("AreaActions", "1.6.5"))
	testza.AssertNoError(t, profile.AddMod("RefinedPower", "3.2.10"))

	serverLocation := os.Getenv("SF_DEDICATED_SERVER")
	if serverLocation != "" {
		testza.AssertNoError(t, os.RemoveAll(filepath.Join(serverLocation, "FactoryGame", "Mods")))

		dir, err := os.ReadDir(filepath.Join(serverLocation, "FactoryGame", "Mods"))
		testza.AssertNoError(t, err)
		spew.Dump(dir)

		installation, err := ctx.Installations.AddInstallation(ctx, "sftp://user:pass@localhost:2222/home/user/server", profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		installation.Vanilla = true
		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
	}

	err = ctx.Wipe()
	testza.AssertNoError(t, err)
}
