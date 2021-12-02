package cmd

import (
	"github.com/satisfactorymodding/ficsit-cli/cli"
	"github.com/satisfactorymodding/ficsit-cli/tea"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cliCmd)
}

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Start interactive CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := cli.InitCLI()
		if err != nil {
			panic(err)
		}

		return tea.RunTea(global)
	},
}
