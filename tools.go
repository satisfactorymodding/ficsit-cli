//go:build tools
// +build tools

package main

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra/doc"

	"github.com/satisfactorymodding/ficsit-cli/cmd"

	_ "github.com/Khan/genqlient/generate"
)

//go:generate go run github.com/Khan/genqlient
//go:generate go run -tags tools tools.go

func main() {
	var err error
	_ = os.RemoveAll("./docs/")

	if err = os.Mkdir("./docs/", 0o777); err != nil {
		panic(err)
	}

	err = doc.GenMarkdownTree(cmd.RootCmd, "./docs/")
	if err != nil {
		panic(err)
	}

	// replace user dir information with generic username
	baseCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}

	var baseLocalDir string

	switch runtime.GOOS {
	case "windows":
		baseLocalDir = os.Getenv("APPDATA")
	case "linux":
		baseLocalDir = filepath.Join(os.Getenv("HOME"), ".local", "share")
	case "darwin":
		baseLocalDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	default:
		panic("unsupported platform: " + runtime.GOOS)
	}

	docFiles, err := os.ReadDir("./docs/")
	if err != nil {
		panic(err)
	}

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	for _, f := range docFiles {
		fPath := "./docs/" + f.Name()
		oldContents, err := os.ReadFile(fPath)
		if err != nil {
			panic(err)
		}

		newContents := strings.ReplaceAll(
			string(oldContents),
			baseCacheDir,
			strings.ReplaceAll(baseCacheDir, user.Username, "{{Username}}"),
		)

		newContents = strings.ReplaceAll(
			newContents,
			baseLocalDir,
			strings.ReplaceAll(baseLocalDir, user.Username, "{{Username}}"),
		)

		os.WriteFile(fPath, []byte(newContents), 0o777)
	}
}
