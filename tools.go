//go:build tools
// +build tools

package main

import (
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/satisfactorymodding/ficsit-cli/cmd"

	_ "github.com/Khan/genqlient/generate"
)

//go:generate go run github.com/Khan/genqlient
//go:generate go run -tags tools tools.go

func main() {
	_ = os.RemoveAll("./docs/")

	if err := os.Mkdir("./docs/", 0o777); err != nil {
		panic(err)
	}

	err := doc.GenMarkdownTree(cmd.RootCmd, "./docs/")
	if err != nil {
		panic(err)
	}
}
