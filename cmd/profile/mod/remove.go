package mod

import (
	"fmt"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove <profile> <mod-reference>",
	Short: "Remove a mod from a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		profile := global.Profiles.GetProfile(args[0])
		if profile == nil {
			return fmt.Errorf("profile with name %s does not exist", args[0])
		}

		profile.RemoveMod(args[1])

		return global.Save()
	},
}
