package mod

import (
	"fmt"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <profile> <mod-reference> [version]",
	Short: "Add mod to a profile",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI()
		if err != nil {
			return err
		}

		version := ">=0.0.0"
		if len(args) > 2 {
			version = args[2]
		}

		profile := global.Profiles.GetProfile(args[0])
		if profile == nil {
			return fmt.Errorf("profile with name %s does not exist", args[0])
		}

		if err := profile.AddMod(args[1], version); err != nil {
			return err
		}

		return global.Save()
	},
}
