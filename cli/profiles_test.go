package cli

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/satisfactorymodding/ficsit-cli/cfg"
)

func init() {
	cfg.SetDefaults()
}

func TestProfilesInit(t *testing.T) {
	profiles, err := InitProfiles()
	testza.AssertNoError(t, err)
	testza.AssertNotNil(t, profiles)
}
