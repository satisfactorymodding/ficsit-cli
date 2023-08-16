//go:build tools
// +build tools

package main

import (
	"os"
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

	docFiles, err := os.ReadDir("./docs/")
	if err != nil {
		panic(err)
	}

	for _, f := range docFiles {
		fPath := "./docs/" + f.Name()
		oldContents, err := os.ReadFile(fPath)
		if err != nil {
			panic(err)
		}

		os.WriteFile(fPath, []byte(strings.ReplaceAll(string(oldContents), baseCacheDir, "<UserCacheDir>")), 0o777)
	}
}
