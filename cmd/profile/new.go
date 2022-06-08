package profile

import (
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI()
		if err != nil {
			return err
		}

		_, err = global.Profiles.AddProfile(args[0])
		if err != nil {
			return err
		}

		return global.Save()
	},
}
