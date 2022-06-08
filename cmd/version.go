package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print current version information",
	Run: func(cmd *cobra.Command, args []string) {
		println(viper.GetString("version"), "-", viper.GetString("commit"))
	},
}
