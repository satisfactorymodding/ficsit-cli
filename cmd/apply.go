package cmd

import (
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"sync"

	"github.com/satisfactorymodding/ficsit-cli/cli"
)

var applyCmd = &cobra.Command{
	Use:   "apply [installation] ...",
	Short: "Apply profiles to all installations",
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		errored := false
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

			wg.Add(1)

			go func(installation *cli.Installation) {
				defer wg.Done()
				if err := installation.Install(global, nil); err != nil {
					errored = true
					slog.Error("installation failed", slog.Any("err", err))
				}
			}(installation)
		}

		wg.Wait()

		if errored {
			os.Exit(1)
		}

		return nil
	},
}
