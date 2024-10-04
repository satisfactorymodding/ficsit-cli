package cli

import (
	"context"
	"time"

	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/cli/provider"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

var _ provider.Provider = (*MockProvider)(nil)

type MockProvider struct {
	resolver.MockProvider
}

func (m MockProvider) Mods(_ context.Context, f ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	if f.Offset > 0 {
		return &ficsit.ModsResponse{
			Mods: ficsit.ModsModsGetMods{
				Count: 5,
				Mods:  []ficsit.ModsModsGetModsModsMod{},
			},
		}, nil
	}

	return &ficsit.ModsResponse{
		Mods: ficsit.ModsModsGetMods{
			Count: 5,
			Mods: []ficsit.ModsModsGetModsModsMod{
				{
					Id:                "9LguyCdDUrpT9N",
					Name:              "Ficsit Remote Monitoring",
					Mod_reference:     "FicsitRemoteMonitoring",
					Last_version_date: time.Now(),
					Created_at:        time.Now(),
				},
				{
					Id:                "DGiLzB3ZErWu2V",
					Name:              "Refined Power",
					Mod_reference:     "RefinedPower",
					Last_version_date: time.Now(),
					Created_at:        time.Now(),
				},
				{
					Id:                "B24emzbs6xVZQr",
					Name:              "RefinedRDLib",
					Mod_reference:     "RefinedRDLib",
					Last_version_date: time.Now(),
					Created_at:        time.Now(),
				},
				{
					Id:                "6vQ6ckVYFiidDh",
					Name:              "Area Actions",
					Mod_reference:     "AreaActions",
					Last_version_date: time.Now(),
					Created_at:        time.Now(),
				},
				{
					Id:                "As2uJmQLLxjXLG",
					Name:              "ModularUI",
					Mod_reference:     "ModularUI",
					Last_version_date: time.Now(),
					Created_at:        time.Now(),
				},
			},
		},
	}, nil
}

var windowsTarget = resolver.Target{
	TargetName: "Windows",
	Link:       "https://api.ficsit.dev/v1/version/7QcfNdo5QAAyoC/Windows/download",
	Hash:       "698df20278b3de3ec30405569a22050c6721cc682389312258c14948bd8f38ae",
}

var windowsServerTarget = resolver.Target{
	TargetName: "WindowsServer",
	Link:       "https://api.ficsit.dev/v1/version/7QcfNdo5QAAyoC/WindowsServer/download",
	Hash:       "7be01ed372e0cf3287a04f5cb32bb9dcf6f6e7a5b7603b7e43669ec4c6c1457f",
}

var linuxServerTarget = resolver.Target{
	TargetName: "LinuxServer",
	Link:       "https://api.ficsit.dev/v1/version/7QcfNdo5QAAyoC/LinuxServer/download",
	Hash:       "bdbd4cb1b472a5316621939ae2fe270fd0e3c0f0a75666a9cbe74ff1313c3663",
}

var commonTargets = []resolver.Target{
	windowsTarget,
	windowsServerTarget,
	linuxServerTarget,
}

func (m MockProvider) ModVersionsWithDependencies(ctx context.Context, modID string) ([]resolver.ModVersion, error) {
	switch modID {
	case "AreaActions":
		return []resolver.ModVersion{
			{
				Version: "1.6.7",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.4.1",
						Optional:  false,
					},
				},
				Targets:          commonTargets,
				RequiredOnRemote: true,
			},
			{
				Version: "1.6.6",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.2.0",
						Optional:  false,
					},
				},
				Targets:          commonTargets,
				RequiredOnRemote: true,
			},
			{
				Version: "1.6.5",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.0.0",
						Optional:  false,
					},
				},
				Targets:          commonTargets,
				RequiredOnRemote: true,
			},
		}, nil
	case "FicsitRemoteMonitoring":
		return []resolver.ModVersion{
			{
				Version: "0.10.1",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.6.0",
						Optional:  false,
					},
				},
				Targets:          commonTargets,
				RequiredOnRemote: true,
			},
			{
				Version: "0.10.0",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.5.0",
						Optional:  false,
					},
				},
				Targets:          commonTargets,
				RequiredOnRemote: true,
			},
			{
				Version: "0.9.8",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.4.1",
						Optional:  false,
					},
				},
				Targets:          commonTargets,
				RequiredOnRemote: true,
			},
		}, nil
	case "ClientOnlyMod":
		return []resolver.ModVersion{
			{
				Version: "0.0.1",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.6.0",
						Optional:  false,
					},
				},
				Targets: []resolver.Target{
					windowsTarget,
				},
				RequiredOnRemote: false,
			},
		}, nil
	case "ServerOnlyMod":
		return []resolver.ModVersion{
			{
				Version: "0.0.1",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.6.0",
						Optional:  false,
					},
				},
				Targets: []resolver.Target{
					windowsServerTarget,
					linuxServerTarget,
				},
				RequiredOnRemote: false,
			},
		}, nil
	case "LaterClientOnlyMod":
		return []resolver.ModVersion{
			{
				Version: "0.0.1",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.6.0",
						Optional:  false,
					},
				},
				Targets:          commonTargets,
				RequiredOnRemote: true,
			},
			{
				Version: "0.0.2",
				Dependencies: []resolver.Dependency{
					{
						ModID:     "SML",
						Condition: "^3.6.0",
						Optional:  false,
					},
				},
				Targets: []resolver.Target{
					windowsTarget,
				},
				RequiredOnRemote: false,
			},
		}, nil
	}

	return m.MockProvider.ModVersionsWithDependencies(ctx, modID) // nolint
}

func (m MockProvider) GetMod(_ context.Context, _ string) (*ficsit.GetModResponse, error) {
	// Currently used only by TUI
	return nil, nil
}

func (m MockProvider) IsOffline() bool {
	return false
}
