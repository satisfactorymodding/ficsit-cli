package installation

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/satisfactorymodding/ficsit-cli/cli"
)

func init() {
	setVanillaCmd.Flags().BoolP("off", "o", false, "Disable vanilla")

	Cmd.AddCommand(setVanillaCmd)
}

var setVanillaCmd = &cobra.Command{
	Use:   "set-vanilla <path>",
	Short: "Change the profile of an installation",
	Args:  cobra.ExactArgs(1),
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

		vanillaOff, err := cmd.Flags().GetBool("off")
		if err != nil {
			return err
		}

		installation.Vanilla = !vanillaOff

		return global.Save()
	},
}
