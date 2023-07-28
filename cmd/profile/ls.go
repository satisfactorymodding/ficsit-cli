package profile

import (
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all profiles",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		for name := range global.Profiles.Profiles {
			println(name)
		}

		return nil
	},
}
