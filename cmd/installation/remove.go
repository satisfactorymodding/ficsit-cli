package installation

import (
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove an installation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		err = global.Installations.DeleteInstallation(args[0])
		if err != nil {
			return err
		}

		return global.Save()
	},
}
