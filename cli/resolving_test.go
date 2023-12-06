package cli

import (
	"math"
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/rs/zerolog/log"

	"github.com/satisfactorymodding/ficsit-cli/cfg"
)

func init() {
	cfg.SetDefaults()
}

func profilesGetResolver() DependencyResolver {
	ctx, err := InitCLI(false)
	if err != nil {
		panic(err)
	}

	return NewDependencyResolver(ctx.Provider)
}

func installWatcher() chan<- InstallUpdate {
	c := make(chan InstallUpdate)
	go func() {
		for i := range c {
			if i.Progress.Total == i.Progress.Completed {
				if i.Type != InstallUpdateTypeOverall {
					log.Info().Str("mod_reference", i.Item.Mod).Str("version", i.Item.Version).Str("type", string(i.Type)).Msg("progress completed")
				} else {
					log.Info().Msg("overall completed")
				}
			} else {
				if i.Type != InstallUpdateTypeOverall {
					if int(i.Progress.Percentage()*100000)%10 == 0 {
						log.Info().Str("mod_reference", i.Item.Mod).Str("version", i.Item.Version).Str("type", string(i.Type)).Float64("percent", i.Progress.Percentage()*100).Msg("progress")
					}
				}
			}
		}
	}()
	return c
}

func TestProfileResolution(t *testing.T) {
	resolver := profilesGetResolver()

	resolved, err := (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"RefinedPower": {
				Version: "3.0.9",
				Enabled: true,
			},
		},
	}).Resolve(resolver, nil, math.MaxInt)

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, resolved, 4)
}

func TestProfileRequiredOlderVersion(t *testing.T) {
	resolver := profilesGetResolver()

	_, err := (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"RefinedPower": {
				Version: "3.0.9",
				Enabled: true,
			},
			"RefinedRDLib": {
				Version: "1.0.6",
				Enabled: true,
			},
		},
	}).Resolve(resolver, nil, math.MaxInt)

	testza.AssertEqual(t, "failed resolving profile dependencies: failed to solve dependencies: Because installing Refined Power (RefinedPower) \"3.0.9\" and Refined Power (RefinedPower) \"3.0.9\" depends on RefinedRDLib \"^1.0.7\", installing RefinedRDLib \"^1.0.7\".\nSo, because installing RefinedRDLib \"1.0.6\", version solving failed.", err.Error())
}

func TestResolutionNonExistentMod(t *testing.T) {
	resolver := profilesGetResolver()

	_, err := (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"ThisModDoesNotExist$$$": {
				Version: ">0.0.0",
				Enabled: true,
			},
		},
	}).Resolve(resolver, nil, math.MaxInt)

	testza.AssertEqual(t, "failed resolving profile dependencies: failed to solve dependencies: failed to make decision: failed to get package versions: mod ThisModDoesNotExist$$$ not found", err.Error())
}

func TestUpdateMods(t *testing.T) {
	ctx, err := InitCLI(false)
	testza.AssertNoError(t, err)

	err = ctx.Wipe()
	testza.AssertNoError(t, err)

	resolver := NewDependencyResolver(ctx.Provider)

	oldLockfile, err := (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"AreaActions": {
				Version: "1.6.5",
				Enabled: true,
			},
		},
	}).Resolve(resolver, nil, math.MaxInt)

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, oldLockfile)
	testza.AssertLen(t, oldLockfile, 2)

	profileName := "UpdateTest"
	profile, err := ctx.Profiles.AddProfile(profileName)
	testza.AssertNoError(t, err)
	testza.AssertNoError(t, profile.AddMod("AreaActions", "<=1.6.6"))

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

		testza.AssertEqual(t, 2, len(*lockFile))
		testza.AssertEqual(t, "1.6.5", (*lockFile)["AreaActions"].Version)

		err = installation.UpdateMods(ctx, []string{"AreaActions"})
		testza.AssertNoError(t, err)

		lockFile, err = installation.LockFile(ctx)
		testza.AssertNoError(t, err)

		testza.AssertEqual(t, 2, len(*lockFile))
		testza.AssertEqual(t, "1.6.6", (*lockFile)["AreaActions"].Version)

		err = installation.Install(ctx, installWatcher())
		testza.AssertNoError(t, err)
	}
}
