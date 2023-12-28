package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"

	"github.com/MarvinJWendt/testza"
	"goftp.io/server/v2"
	"goftp.io/server/v2/driver/file"

	"github.com/satisfactorymodding/ficsit-cli/cfg"
)

// NOTE:
//
// This code contains sleep.
// This is because github actions are special.
// They don't properly sync to disk.
// And Go is faster than their disk.
// So tests are flaky :)
// DO NOT REMOVE THE SLEEP!

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
		time.Sleep(time.Second)
		testza.AssertNoError(t, os.RemoveAll(filepath.Join(serverLocation, "FactoryGame", "Mods")))
		time.Sleep(time.Second)

		installation, err := ctx.Installations.AddInstallation(ctx, serverLocation, profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		installation.Vanilla = true
		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
		time.Sleep(time.Second)
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
		driver, err := file.NewDriver(serverLocation)
		testza.AssertNoError(t, err)

		s, err := server.NewServer(&server.Options{
			Driver: driver,
			Auth: &server.SimpleAuth{
				Name:     "user",
				Password: "pass",
			},
			Port: 2121,
			Perm: server.NewSimplePerm("root", "root"),
		})
		testza.AssertNoError(t, err)
		defer testza.AssertNoError(t, s.Shutdown())

		go func() {
			testza.AssertNoError(t, s.ListenAndServe())
		}()

		time.Sleep(time.Second)
		testza.AssertNoError(t, os.RemoveAll(filepath.Join(serverLocation, "FactoryGame", "Mods")))
		time.Sleep(time.Second)

		installation, err := ctx.Installations.AddInstallation(ctx, "ftp://user:pass@localhost:2121/", profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		installation.Vanilla = true
		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
		time.Sleep(time.Second)
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
		time.Sleep(time.Second)
		testza.AssertNoError(t, os.RemoveAll(filepath.Join(serverLocation, "FactoryGame", "Mods")))
		time.Sleep(time.Second)

		installation, err := ctx.Installations.AddInstallation(ctx, "sftp://user:pass@localhost:2222/home/user/server", profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		installation.Vanilla = true
		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
		time.Sleep(time.Second)
	}

	err = ctx.Wipe()
	testza.AssertNoError(t, err)
}
