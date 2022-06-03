package cmd

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pkg/errors"

	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "ficsit",
	Short: "cli mod manager for satisfactory",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

		writers := make([]io.Writer, 0)
		if viper.GetBool("pretty") {
			pterm.EnableStyling()
		} else {
			pterm.DisableStyling()
		}

		if !viper.GetBool("quiet") {
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			})
		}

		if viper.GetString("log-file") != "" {
			logFile, err := os.OpenFile(viper.GetString("log-file"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)

			if err != nil {
				return errors.Wrap(err, "failed to open log file")
			}

			writers = append(writers, logFile)
		}

		log.Logger = zerolog.New(io.MultiWriter(writers...)).With().Timestamp().Logger()

		return nil
	},
}

func Execute() {
	// Execute tea as default
	cmd, _, err := rootCmd.Find(os.Args[1:])

	// Allow opening via explorer
	cobra.MousetrapHelpText = ""

	cli := len(os.Args) >= 2 && os.Args[1] == "cli"
	if (len(os.Args) <= 1 || os.Args[1] != "help") && (err != nil || cmd == rootCmd) {
		args := append([]string{"cli"}, os.Args[1:]...)
		rootCmd.SetArgs(args)
		cli = true
	}

	// Always be quiet in CLI mode
	if cli {
		viper.Set("quiet", true)
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	var baseLocalDir string

	switch runtime.GOOS {
	case "windows":
		baseLocalDir = os.Getenv("APPDATA")
	case "linux":
		baseLocalDir = path.Join(os.Getenv("HOME"), ".local", "share")
	default:
		panic("unsupported platform: " + runtime.GOOS)
	}

	viper.Set("base-local-dir", baseLocalDir)

	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().String("log", "info", "The log level to output")
	rootCmd.PersistentFlags().String("log-file", "", "File to output logs to")
	rootCmd.PersistentFlags().Bool("quiet", false, "Do not log anything to console")
	rootCmd.PersistentFlags().Bool("pretty", true, "Whether to render pretty terminal output")

	rootCmd.PersistentFlags().Bool("dry-run", false, "Dry-run. Do not save any changes")

	rootCmd.PersistentFlags().String("cache-dir", filepath.Clean(filepath.Join(baseCacheDir, "ficsit")), "The cache directory")
	rootCmd.PersistentFlags().String("local-dir", filepath.Clean(filepath.Join(baseLocalDir, "ficsit")), "The local directory")
	rootCmd.PersistentFlags().String("profiles-file", "profiles.json", "The profiles file")
	rootCmd.PersistentFlags().String("installations-file", "installations.json", "The installations file")

	rootCmd.PersistentFlags().String("api-base", "https://api.ficsit.app", "URL for API")
	rootCmd.PersistentFlags().String("graphql-api", "/v2/query", "Path for GraphQL API")

	_ = viper.BindPFlag("log", rootCmd.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("pretty", rootCmd.PersistentFlags().Lookup("pretty"))

	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))

	_ = viper.BindPFlag("cache-dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
	_ = viper.BindPFlag("local-dir", rootCmd.PersistentFlags().Lookup("local-dir"))
	_ = viper.BindPFlag("profiles-file", rootCmd.PersistentFlags().Lookup("profiles-file"))
	_ = viper.BindPFlag("installations-file", rootCmd.PersistentFlags().Lookup("installations-file"))

	_ = viper.BindPFlag("api-base", rootCmd.PersistentFlags().Lookup("api-base"))
	_ = viper.BindPFlag("graphql-api", rootCmd.PersistentFlags().Lookup("graphql-api"))
}
