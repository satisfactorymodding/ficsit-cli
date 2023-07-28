package cmd

import (
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [installation] ...",
	Short: "Apply profiles to all installations",
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		for _, installation := range global.Installations.Installations {
			if len(args) > 0 {
				found := false

				for _, installPath := range args {
					if installation.Path == installPath {
						found = true
						break
					}
				}

				if !found {
					continue
				}
			}

			if err := installation.Install(global, nil); err != nil {
				return err
			}
		}

		return nil
	},
}
