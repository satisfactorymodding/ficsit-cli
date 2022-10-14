package installation

import (
	"github.com/spf13/cobra"

	"github.com/satisfactorymodding/ficsit-cli/cli"
)

func init() {
	Cmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <path> [profile]",
	Short: "Add an installation",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		profile := global.Profiles.SelectedProfile
		if len(args) > 1 {
			profile = args[1]
		}

		_, err = global.Installations.AddInstallation(global, args[0], profile)
		if err != nil {
			return err
		}

		return global.Save()
	},
}
