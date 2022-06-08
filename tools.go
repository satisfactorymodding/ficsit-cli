//go:build tools
// +build tools

package main

import (
	_ "github.com/Khan/genqlient/generate"
	"github.com/satisfactorymodding/ficsit-cli/cmd"
	"github.com/spf13/cobra/doc"
	"os"
)

//go:generate go run -tags tools tools.go
//go:generate go run github.com/Khan/genqlient

func main() {
	_ = os.RemoveAll("./docs/")

	if err := os.Mkdir("./docs/", 0777); err != nil {
		panic(err)
	}

	err := doc.GenMarkdownTree(cmd.RootCmd, "./docs/")
	if err != nil {
		panic(err)
	}
}
