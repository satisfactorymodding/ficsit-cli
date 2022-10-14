package profile

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/satisfactorymodding/ficsit-cli/cli"
)

func init() {
	Cmd.AddCommand(modsCmd)
}

var modsCmd = &cobra.Command{
	Use:   "mods <profile>",
	Short: "List all mods in a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		profile := global.Profiles.GetProfile(args[0])
		if profile == nil {
			return errors.New("profile not found")
		}

		for reference, mod := range profile.Mods {
			println(reference, mod.Version)
		}

		return nil
	},
}
