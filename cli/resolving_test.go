package cli

import (
	"context"
	"math"
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"
	resolver "github.com/satisfactorymodding/ficsit-resolver"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cfg"
)

func init() {
	cfg.SetDefaults()
}

func TestUpdateMods(t *testing.T) {
	ctx, err := InitCLI(false)
	testza.AssertNoError(t, err)

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	ctx.Provider = MockProvider{}

	depResolver := resolver.NewDependencyResolver(ctx.Provider, viper.GetString("api-base"))

	oldLockfile, err := depResolver.ResolveModDependencies(context.Background(), map[string]string{
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

		err = installation.Install(ctx, nil)
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

		err = installation.Install(ctx, nil)
		testza.AssertNoError(t, err)
	}
}
