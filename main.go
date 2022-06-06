package main

import "github.com/satisfactorymodding/ficsit-cli/cmd"

var (
	version = "dev"
	commit  = "none"
)

func main() {
	cmd.Execute(version, commit)
}
