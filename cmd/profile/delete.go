package profile

import (
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		err = global.Profiles.DeleteProfile(args[0])
		if err != nil {
			return err
		}

		return global.Save()
	},
}
