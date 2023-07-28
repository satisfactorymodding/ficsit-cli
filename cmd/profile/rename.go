package profile

import (
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(renameCmd)
}

var renameCmd = &cobra.Command{
	Use:   "rename <old> <name>",
	Short: "Rename a profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		err = global.Profiles.RenameProfile(global, args[0], args[1])
		if err != nil {
			return err
		}

		return global.Save()
	},
}
