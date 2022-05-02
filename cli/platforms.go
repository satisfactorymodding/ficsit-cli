package cli

import "path"

type Platform struct {
	VersionPath  string
	LockfilePath string
}

var platforms = []Platform{
	{
		VersionPath:  path.Join("Engine", "Binaries", "Linux", "UE4Server-Linux-Shipping.version"),
		LockfilePath: path.Join("FactoryGame", "Mods", "mods-lock.json"),
	},
	{
		VersionPath:  path.Join("Engine", "Binaries", "Win64", "UE4Server-Win64-Shipping.version"),
		LockfilePath: path.Join("FactoryGame", "Mods", "mods-lock.json"),
	},
	{
		VersionPath:  path.Join("Engine", "Binaries", "Win64", "FactoryGame-Win64-Shipping.version"),
		LockfilePath: path.Join("FactoryGame", "Mods", "mods-lock.json"),
	},
}
