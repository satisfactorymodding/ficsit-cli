package cli

import "path/filepath"

type Platform struct {
	VersionPath  string
	LockfilePath string
}

var platforms = []Platform{
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Linux", "UE4Server-Linux-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "UE4Server-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "FactoryGame-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
	},
}
