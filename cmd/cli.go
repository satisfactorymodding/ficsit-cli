package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea"
)

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Start interactive CLI (default)",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().
			Str("version", viper.GetString("version")).
			Str("commit", viper.GetString("commit")).
			Msg("initialized")

		global, err := cli.InitCLI(false)
		if err != nil {
			return err
		}

		return tea.RunTea(global)
	},
}
