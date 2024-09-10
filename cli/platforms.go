package cli

import "path/filepath"

type Platform struct {
	VersionPath  string
	LockfilePath string
	TargetName   string
}

var platforms = []Platform{
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Linux", "UnrealServer-Linux-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "LinuxServer",
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "UnrealServer-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "WindowsServer",
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "FactoryGame-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "Windows",
	},
	// Update 9 stuff below
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Linux", "FactoryServer-Linux-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "LinuxServer",
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "FactoryServer-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "WindowsServer",
	},
	// 1.0 stuff
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "FactoryGameSteam-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "Windows",
	},
	{
		VersionPath:  filepath.Join("Engine", "Binaries", "Win64", "FactoryGameEGS-Win64-Shipping.version"),
		LockfilePath: filepath.Join("FactoryGame", "Mods"),
		TargetName:   "Windows",
	},
}
