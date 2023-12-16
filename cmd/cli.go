package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea"
)

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Start interactive CLI (default)",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info(
			"interactive cli initialized",
			slog.String("version", viper.GetString("version")),
			slog.String("commit", viper.GetString("commit")),
		)

		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		return tea.RunTea(global)
	},
}
