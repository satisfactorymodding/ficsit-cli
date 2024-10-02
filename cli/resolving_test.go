package cli

import (
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/cfg"
)

func init() {
	cfg.SetDefaults()
}

func installWatcher() chan<- InstallUpdate {
	c := make(chan InstallUpdate)
	go func() {
		for i := range c {
			if i.Progress.Total == i.Progress.Completed {
				if i.Type != InstallUpdateTypeOverall {
					slog.Info("progress completed", slog.String("mod_reference", i.Item.Mod), slog.String("version", i.Item.Version), slog.Any("type", i.Type))
				} else {
					slog.Info("overall completed")
				}
			}
		}
	}()
	return c
}

func TestClientOnlyMod(t *testing.T) {
	ctx, err := InitCLI(false)
	testza.AssertNoError(t, err)

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	ctx.Provider = MockProvider{}

	profileName := "ClientOnlyModTest"
	profile, err := ctx.Profiles.AddProfile(profileName)
	profile.RequiredTargets = []resolver.TargetName{resolver.TargetNameWindows, resolver.TargetNameWindowsServer, resolver.TargetNameLinuxServer}
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, profile.AddMod("ClientOnlyMod", "<=0.0.1"))

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
	}
}

func TestServerOnlyMod(t *testing.T) {
	ctx, err := InitCLI(false)
	testza.AssertNoError(t, err)

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	ctx.Provider = MockProvider{}

	profileName := "ServerOnlyModTest"
	profile, err := ctx.Profiles.AddProfile(profileName)
	profile.RequiredTargets = []resolver.TargetName{resolver.TargetNameWindows, resolver.TargetNameWindowsServer, resolver.TargetNameLinuxServer}
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, profile.AddMod("ServerOnlyMod", "<=0.0.1"))

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
	}
}

func TestRemoveWhenNotSupported(t *testing.T) {
	ctx, err := InitCLI(false)
	testza.AssertNoError(t, err)

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	ctx.Provider = MockProvider{}

	profileName := "ClientOnlyModTest"
	profile, err := ctx.Profiles.AddProfile(profileName)
	profile.RequiredTargets = []resolver.TargetName{resolver.TargetNameWindows, resolver.TargetNameWindowsServer, resolver.TargetNameLinuxServer}
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, profile.AddMod("LaterClientOnlyMod", "0.0.1"))

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

		_, err = os.Stat(filepath.Join(serverLocation, "FactoryGame", "Mods", "LaterClientOnlyMod"))
		testza.AssertNoError(t, err)

		testza.AssertNoError(t, profile.AddMod("LaterClientOnlyMod", "0.0.2"))

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		_, err = os.Stat(filepath.Join(serverLocation, "FactoryGame", "Mods", "LaterClientOnlyMod"))
		testza.AssertNotNil(t, err)
		testza.AssertErrorIs(t, err, os.ErrNotExist)
	}
}

func TestUpdateMods(t *testing.T) {
	ctx, err := InitCLI(false)
	testza.AssertNoError(t, err)

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	ctx.Provider = MockProvider{}

	depResolver := resolver.NewDependencyResolver(ctx.Provider)

	oldLockfile, err := depResolver.ResolveModDependencies(map[string]string{
		"FicsitRemoteMonitoring": "0.9.8",
	}, nil, math.MaxInt, nil)

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, oldLockfile)
	testza.AssertLen(t, oldLockfile.Mods, 2)

	profileName := "UpdateTest"
	profile, err := ctx.Profiles.AddProfile(profileName)
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, profile.AddMod("FicsitRemoteMonitoring", "<=0.10.0"))

	serverLocation := os.Getenv("SF_DEDICATED_SERVER")
	if serverLocation != "" {
		installation, err := ctx.Installations.AddInstallation(ctx, serverLocation, profileName)
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, installation)

		err = installation.WriteLockFile(ctx, oldLockfile)
		testza.AssertNoError(t, err)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)

		lockFile, err := installation.LockFile(ctx)
		testza.AssertNoError(t, err)

		testza.AssertEqual(t, 2, len(lockFile.Mods))
		testza.AssertEqual(t, "0.9.8", (lockFile.Mods)["FicsitRemoteMonitoring"].Version)

		err = installation.UpdateMods(ctx, []string{"FicsitRemoteMonitoring"})
		testza.AssertNoError(t, err)

		lockFile, err = installation.LockFile(ctx)
		testza.AssertNoError(t, err)

		testza.AssertEqual(t, 2, len(lockFile.Mods))
		testza.AssertEqual(t, "0.10.0", (lockFile.Mods)["FicsitRemoteMonitoring"].Version)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
	}
}
