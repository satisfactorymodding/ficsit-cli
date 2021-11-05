package cmd

import (
	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/tea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "ficsit",
	Short: "cli mod manager for satisfactory",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.SetEnvPrefix("ficsit")
		viper.AutomaticEnv()

		_ = viper.ReadInConfig()

		level, err := zerolog.ParseLevel(viper.GetString("log"))
		if err != nil {
			panic(err)
		}

		zerolog.SetGlobalLevel(level)

		if viper.GetBool("pretty") {
			pterm.EnableStyling()
			log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			})
		} else {
			pterm.DisableStyling()
		}
	},
}

func Execute() {
	// Execute tea as default
	cmd, _, err := rootCmd.Find(os.Args[1:])
	if (len(os.Args) <= 1 || os.Args[1] != "help") && (err != nil || cmd == rootCmd) {
		tea.RunTea()
		return
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().String("log", "info", "The log level to output")
	rootCmd.PersistentFlags().Bool("pretty", true, "Whether to render pretty terminal output")

	rootCmd.PersistentFlags().String("cache-dir", path.Join(baseCacheDir, "ficsit"), "The cache directory")

	_ = viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("pretty", rootCmd.PersistentFlags().Lookup("pretty"))

	_ = viper.BindPFlag("cache-dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
}
