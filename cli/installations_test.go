package cli

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/avast/retry-go"

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

	err := retry.Do(func() error {
		ctx, err := InitCLI(false)
		if err != nil {
			return err
		}

		err = ctx.Wipe()
		if err != nil {
			return err
		}

		err = ctx.ReInit()
		if err != nil {
			return err
		}

		ctx.Provider = MockProvider{}

		profileName := "InstallationTest"
		profile, err := ctx.Profiles.AddProfile(profileName)
		if err != nil {
			return err
		}

		testza.AssertNoError(t, profile.AddMod("AreaActions", "1.6.5"))
		testza.AssertNoError(t, profile.AddMod("RefinedPower", "3.2.10"))

		serverLocation := os.Getenv("SF_DEDICATED_SERVER")
		if serverLocation != "" {
			time.Sleep(time.Second)

			testza.AssertNoError(t, os.RemoveAll(filepath.Join(serverLocation, "FactoryGame", "Mods")))
			time.Sleep(time.Second)

			installation, err := ctx.Installations.AddInstallation(ctx, "ftp://user:pass@localhost:2121/server", profileName)
			if err != nil {
				return err
			}

			testza.AssertNotNil(t, installation)

			err = installation.Install(ctx, installWatcher())
			if err != nil {
				return err
			}

			installation.Vanilla = true
			err = installation.Install(ctx, installWatcher())
			if err != nil {
				return err
			}
			testza.AssertNoError(t, err)

			time.Sleep(time.Second)
		}

		err = ctx.Wipe()
		if err != nil {
			return err
		}

		return nil
	},
		retry.Attempts(30),
		retry.Delay(time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(n uint, err error) {
			if n > 0 {
				slog.Info("retrying ftp test", slog.Uint64("n", uint64(n)))
			}
		}),
	)
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
