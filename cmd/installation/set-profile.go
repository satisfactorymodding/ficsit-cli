package installation

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/satisfactorymodding/ficsit-cli/cli"
)

func init() {
	Cmd.AddCommand(setProfileCmd)
}

var setProfileCmd = &cobra.Command{
	Use:   "set-profile <path> <profile>",
	Short: "Change the profile of an installation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		var installation *cli.Installation
		for _, install := range global.Installations.Installations {
			if install.Path == args[0] {
				installation = install
				break
			}
		}

		if installation == nil {
			return errors.New("installation not found")
		}

		err = installation.SetProfile(global, args[1])
		if err != nil {
			return err
		}

		return global.Save()
	},
}
