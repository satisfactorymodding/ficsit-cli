package cli

import "path/filepath"

type Platform struct {
	VersionPath  string
	LockfilePath string
	TargetName   string
}

var platforms = []Platform{
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Linux", "UE4Server-Linux-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "LinuxServer",
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "UE4Server-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "WindowsServer",
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "FactoryGame-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "WindowsNoEditor", // TODO: Support both WindowsNoEditor (UE4) and Windows (UE5)
	},
}
