package cfg

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

func SetDefaults() {
	_, file, _, _ := runtime.Caller(0)
	viper.SetDefault("cache-dir", filepath.Clean(filepath.Join(filepath.Dir(file), "../", "testdata", "cache")))
	viper.SetDefault("local-dir", filepath.Clean(filepath.Join(filepath.Dir(file), "../", "testdata", "local")))
	viper.SetDefault("base-local-dir", filepath.Clean(filepath.Join(filepath.Dir(file), "../", "testdata")))
	viper.SetDefault("profiles-file", "profiles.json")
	viper.SetDefault("installations-file", "installations.json")
	viper.SetDefault("dry-run", false)
	viper.SetDefault("api-base", "https://api.ficsit.dev")
	viper.SetDefault("graphql-api", "/v2/query")
	viper.SetDefault("concurrent-downloads", 5)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}
