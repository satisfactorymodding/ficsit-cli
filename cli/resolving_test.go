package cli

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

func TestProfileResolution(t *testing.T) {
	api := ficsit.InitAPI()
	resolver := NewDependencyResolver(api)

	resolved, err := (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"RefinedPower": {
				Version: "3.0.9",
			},
		},
	}).Resolve(resolver, nil)

	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, resolved)
	testza.AssertLen(t, *resolved, 3)

	_, err = (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"RefinedPower": {
				Version: "3.0.9",
			},
			"RefinedRDLib": {
				Version: "1.0.6",
			},
		},
	}).Resolve(resolver, nil)

	testza.AssertEqual(t, "failed resolving profile dependencies: mod RefinedRDLib version 1.0.6 does not match constraint ^1.0.7", err.Error())

	_, err = (&Profile{
		Name: DefaultProfileName,
		Mods: map[string]ProfileMod{
			"ThisModDoesNotExist$$$": {
				Version: ">0.0.0",
			},
		},
	}).Resolve(resolver, nil)

	testza.AssertEqual(t, "failed resolving profile dependencies: failed resolving dependency: ThisModDoesNotExist$$$", err.Error())
}
