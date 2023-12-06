package cli

import (
	"math"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func profilesGetResolver() DependencyResolver {
	ctx, err := InitCLI(false)
	if err != nil {
		panic(err)
	}

	return NewDependencyResolver(ctx.Provider)
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
