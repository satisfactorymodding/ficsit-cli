package profile

import (
	"github.com/spf13/cobra"

	"github.com/satisfactorymodding/ficsit-cli/cli"
)

func init() {
	Cmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
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
