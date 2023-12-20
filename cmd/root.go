package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/lmittmann/tint"
	"github.com/pterm/pterm"
	slogmulti "github.com/samber/slog-multi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cmd/installation"
	"github.com/satisfactorymodding/ficsit-cli/cmd/mod"
	"github.com/satisfactorymodding/ficsit-cli/cmd/profile"
	"github.com/satisfactorymodding/ficsit-cli/cmd/smr"
)

var RootCmd = &cobra.Command{
	Use:   "ficsit",
	Short: "cli mod manager for satisfactory",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.SetEnvPrefix("ficsit")
		viper.AutomaticEnv()

		_ = viper.ReadInConfig()

		handlers := make([]slog.Handler, 0)
		if viper.GetBool("pretty") {
			pterm.EnableStyling()
		} else {
			pterm.DisableStyling()
		}

		const (
			ansiReset         = "\033[0m"
			ansiBold          = "\033[1m"
			ansiWhite         = "\033[38m"
			ansiBrightMagenta = "\033[95m"
		)

		level := slog.LevelInfo
		if err := (&level).UnmarshalText([]byte(viper.GetString("log"))); err != nil {
			return fmt.Errorf("failed parsing level: %w", err)
		}

		if !viper.GetBool("quiet") {
			handlers = append(handlers, tint.NewHandler(os.Stdout, &tint.Options{
				Level:      level,
				AddSource:  true,
				TimeFormat: time.RFC3339Nano,
				ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
					if attr.Key == slog.LevelKey {
						level := attr.Value.Any().(slog.Level)
						if level == slog.LevelDebug {
							attr.Value = slog.StringValue(ansiBrightMagenta + "DBG" + ansiReset)
						}
					} else if attr.Key == slog.MessageKey {
						attr.Value = slog.StringValue(ansiBold + ansiWhite + fmt.Sprint(attr.Value.Any()) + ansiReset)
					}
					return attr
				},
			}))
		}

		if viper.GetString("log-file") != "" {
			logFile, err := os.OpenFile(viper.GetString("log-file"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o777)
			if err != nil {
				return fmt.Errorf("failed to open log file: %w", err)
			}

			handlers = append(handlers, slog.NewJSONHandler(logFile, &slog.HandlerOptions{}))
		}

		slog.SetDefault(slog.New(
			slogmulti.Fanout(handlers...),
		))

		return nil
	},
}

func Execute(version string, commit string) {
	// Execute tea as default
	cmd, _, err := RootCmd.Find(os.Args[1:])

	// Allow opening via explorer
	cobra.MousetrapHelpText = ""

	cli := len(os.Args) >= 2 && os.Args[1] == "cli"
	if (len(os.Args) <= 1 || (os.Args[1] != "help" && os.Args[1] != "--help" && os.Args[1] != "-h")) && (err != nil || cmd == RootCmd) {
		args := append([]string{"cli"}, os.Args[1:]...)
		RootCmd.SetArgs(args)
		cli = true
	}

	// Always be quiet in CLI mode
	if cli {
		viper.Set("quiet", true)
	}

	viper.Set("version", version)
	viper.Set("commit", commit)

	if err := RootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(cliCmd)
	RootCmd.AddCommand(applyCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(searchCmd)
	RootCmd.AddCommand(profile.Cmd)
	RootCmd.AddCommand(installation.Cmd)
	RootCmd.AddCommand(mod.Cmd)
	RootCmd.AddCommand(smr.Cmd)

	var baseLocalDir string

	switch runtime.GOOS {
	case "windows":
		baseLocalDir = os.Getenv("APPDATA")
	case "linux":
		baseLocalDir = filepath.Join(os.Getenv("HOME"), ".local", "share")
	default:
		panic("unsupported platform: " + runtime.GOOS)
	}

	viper.Set("base-local-dir", baseLocalDir)

	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	RootCmd.PersistentFlags().String("log", "info", "The log level to output")
	RootCmd.PersistentFlags().String("log-file", "", "File to output logs to")
	RootCmd.PersistentFlags().Bool("quiet", false, "Do not log anything to console")
	RootCmd.PersistentFlags().Bool("pretty", true, "Whether to render pretty terminal output")

	RootCmd.PersistentFlags().Bool("dry-run", false, "Dry-run. Do not save any changes")

	RootCmd.PersistentFlags().String("cache-dir", filepath.Clean(filepath.Join(baseCacheDir, "ficsit")), "The cache directory")
	RootCmd.PersistentFlags().String("local-dir", filepath.Clean(filepath.Join(baseLocalDir, "ficsit")), "The local directory")
	RootCmd.PersistentFlags().String("profiles-file", "profiles.json", "The profiles file")
	RootCmd.PersistentFlags().String("installations-file", "installations.json", "The installations file")

	RootCmd.PersistentFlags().String("api-base", "https://api.ficsit.app", "URL for API")
	RootCmd.PersistentFlags().String("graphql-api", "/v2/query", "Path for GraphQL API")
	RootCmd.PersistentFlags().String("api-key", "", "API key to use when sending requests")

	RootCmd.PersistentFlags().Bool("offline", false, "Whether to only use local data")
	RootCmd.PersistentFlags().Int("concurrent-downloads", 5, "Maximum number of concurrent downloads")

	_ = viper.BindPFlag("log", RootCmd.PersistentFlags().Lookup("log"))
	_ = viper.BindPFlag("log-file", RootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("quiet", RootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("pretty", RootCmd.PersistentFlags().Lookup("pretty"))

	_ = viper.BindPFlag("dry-run", RootCmd.PersistentFlags().Lookup("dry-run"))

	_ = viper.BindPFlag("cache-dir", RootCmd.PersistentFlags().Lookup("cache-dir"))
	_ = viper.BindPFlag("local-dir", RootCmd.PersistentFlags().Lookup("local-dir"))
	_ = viper.BindPFlag("profiles-file", RootCmd.PersistentFlags().Lookup("profiles-file"))
	_ = viper.BindPFlag("installations-file", RootCmd.PersistentFlags().Lookup("installations-file"))

	_ = viper.BindPFlag("api-base", RootCmd.PersistentFlags().Lookup("api-base"))
	_ = viper.BindPFlag("graphql-api", RootCmd.PersistentFlags().Lookup("graphql-api"))
	_ = viper.BindPFlag("api-key", RootCmd.PersistentFlags().Lookup("api-key"))

	_ = viper.BindPFlag("offline", RootCmd.PersistentFlags().Lookup("offline"))
	_ = viper.BindPFlag("concurrent-downloads", RootCmd.PersistentFlags().Lookup("concurrent-downloads"))
}
