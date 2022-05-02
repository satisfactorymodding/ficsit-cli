package cli

import (
	"math"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestProfileResolution(t *testing.T) {
	ctx, err := InitCLI()
	testza.AssertNoError(t, err)

	resolver := NewDependencyResolver(ctx.APIClient)

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

	_, err = (&Profile{
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

	testza.AssertEqual(t, "failed resolving profile dependencies: failed resolving dependencies. requires different versions of RefinedRDLib", err.Error())

	_, err = (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"ThisModDoesNotExist$$$": {
				Version: ">0.0.0",
				Enabled: true,
			},
		},
	}).Resolve(resolver, nil, math.MaxInt)

	testza.AssertEqual(t, "failed resolving profile dependencies: failed resolving dependency: ThisModDoesNotExist$$$", err.Error())
}
