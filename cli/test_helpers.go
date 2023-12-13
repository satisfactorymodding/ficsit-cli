package cli

import (
	"context"
	"time"

	"github.com/satisfactorymodding/ficsit-cli/cli/provider"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

var _ provider.Provider = (*MockProvider)(nil)

type MockProvider struct{}

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

func (m MockProvider) GetMod(_ context.Context, _ string) (*ficsit.GetModResponse, error) {
	// Currently used only by TUI
	return nil, nil
}

func (m MockProvider) ModVersions(_ context.Context, modReference string, _ ficsit.VersionFilter) (*ficsit.ModVersionsResponse, error) {
	switch modReference {
	//nolint
	case "RefinedPower":
		return &ficsit.ModVersionsResponse{Mod: ficsit.ModVersionsMod{
			Id: "DGiLzB3ZErWu2V",
			Versions: []ficsit.ModVersionsModVersionsVersion{
				{Id: "Eqgr4VcB8y1z9a", Version: "3.2.13"},
				{Id: "BwVKMJNP8doDLg", Version: "3.2.11"},
				{Id: "4XTjMpqFngbu9r", Version: "3.2.10"},
			},
		}}, nil
	//nolint
	case "RefinedRDLib":
		return &ficsit.ModVersionsResponse{Mod: ficsit.ModVersionsMod{
			Id: "B24emzbs6xVZQr",
			Versions: []ficsit.ModVersionsModVersionsVersion{
				{Id: "2XcE6RUzGhZW7p", Version: "1.1.7"},
				{Id: "52RMLEigqT5Ksn", Version: "1.1.6"},
				{Id: "F4HY9eP4D5XjWQ", Version: "1.1.5"},
			},
		}}, nil
	}

	panic("ModVersions: " + modReference)
}

func (m MockProvider) SMLVersions(_ context.Context) (*ficsit.SMLVersionsResponse, error) {
	return &ficsit.SMLVersionsResponse{
		SmlVersions: ficsit.SMLVersionsSmlVersionsGetSMLVersions{
			Count: 4,
			Sml_versions: []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion{
				{
					Id:                   "v2.2.1",
					Version:              "2.2.1",
					Satisfactory_version: 125236,
					Targets:              []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersionTargetsSMLVersionTarget{},
				},
				{
					Id:                   "v3.3.2",
					Version:              "3.3.2",
					Satisfactory_version: 194714,
					Targets: []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersionTargetsSMLVersionTarget{
						{
							TargetName: ficsit.TargetNameWindows,
							Link:       "https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v3.3.2/SML.zip",
						},
					},
				},
				{
					Id:                   "v3.6.0",
					Version:              "3.6.0",
					Satisfactory_version: 264901,
					Targets: []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersionTargetsSMLVersionTarget{
						{
							TargetName: ficsit.TargetNameWindows,
							Link:       "https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v3.6.0/SML.zip",
						},
						{
							TargetName: ficsit.TargetNameWindowsserver,
							Link:       "https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v3.6.0/SML.zip",
						},
						{
							TargetName: ficsit.TargetNameLinuxserver,
							Link:       "https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v3.6.0/SML.zip",
						},
					},
				},
				{
					Id:                   "v3.6.1",
					Version:              "3.6.1",
					Satisfactory_version: 264901,
					Targets: []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersionTargetsSMLVersionTarget{
						{
							TargetName: ficsit.TargetNameWindows,
							Link:       "https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v3.6.1/SML.zip",
						},
						{
							TargetName: ficsit.TargetNameWindowsserver,
							Link:       "https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v3.6.1/SML.zip",
						},
						{
							TargetName: ficsit.TargetNameLinuxserver,
							Link:       "https://github.com/satisfactorymodding/SatisfactoryModLoader/releases/download/v3.6.1/SML.zip",
						},
					},
				},
			},
		},
	}, nil
}

var commonTargets = []ficsit.Target{
	{
		TargetName: "Windows",
		Hash:       "62f5c84eca8480b3ffe7d6c90f759e3b463f482530e27d854fd48624fdd3acc9",
	},
	{
		TargetName: "WindowsServer",
		Hash:       "8a83fcd4abece4192038769cc672fff6764d72c32fb6c7a8c58d66156bb07917",
	},
	{
		TargetName: "LinuxServer",
		Hash:       "8739c76e681f900923b900c9df0ef75cf421d39cabb54650c4b9ad19b6a76d85",
	},
}

func (m MockProvider) ModVersionsWithDependencies(_ context.Context, modID string) (*ficsit.AllVersionsResponse, error) {
	switch modID {
	case "RefinedPower":
		return &ficsit.AllVersionsResponse{
			Success: true,
			Data: []ficsit.ModVersion{
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "3.2.13",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "ModularUI",
							Condition: "^2.1.11",
							Optional:  false,
						},
						{
							ModID:     "RefinedRDLib",
							Condition: "^1.1.7",
							Optional:  false,
						},
						{
							ModID:     "SML",
							Condition: "^3.6.1",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "3.2.11",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "ModularUI",
							Condition: "^2.1.10",
							Optional:  false,
						},
						{
							ModID:     "RefinedRDLib",
							Condition: "^1.1.6",
							Optional:  false,
						},
						{
							ModID:     "SML",
							Condition: "^3.6.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "3.2.10",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "ModularUI",
							Condition: "^2.1.9",
							Optional:  false,
						},
						{
							ModID:     "RefinedRDLib",
							Condition: "^1.1.5",
							Optional:  false,
						},
						{
							ModID:     "SML",
							Condition: "^3.6.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
			},
		}, nil
	case "AreaActions":
		return &ficsit.AllVersionsResponse{
			Success: true,
			Data: []ficsit.ModVersion{
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "1.6.7",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.4.1",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "1.6.6",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.2.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "1.6.5",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.0.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
			},
		}, nil
	case "RefinedRDLib":
		return &ficsit.AllVersionsResponse{
			Success: true,
			Data: []ficsit.ModVersion{
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "1.1.7",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.6.1",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "1.1.6",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.6.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "1.1.5",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.6.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
			},
		}, nil
	case "ModularUI":
		return &ficsit.AllVersionsResponse{
			Success: true,
			Data: []ficsit.ModVersion{
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "2.1.12",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.6.1",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "2.1.11",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.6.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "2.1.10",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.6.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
			},
		}, nil
	case "ThisModDoesNotExist$$$":
		return &ficsit.AllVersionsResponse{
			Success: false,
			Error: &ficsit.Error{
				Message: "mod not found",
				Code:    200,
			},
		}, nil
	case "FicsitRemoteMonitoring":
		return &ficsit.AllVersionsResponse{
			Success: true,
			Data: []ficsit.ModVersion{
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "0.10.1",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.6.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "0.10.0",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.5.0",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
				{
					ID:      "7QcfNdo5QAAyoC",
					Version: "0.9.8",
					Dependencies: []ficsit.Dependency{
						{
							ModID:     "SML",
							Condition: "^3.4.1",
							Optional:  false,
						},
					},
					Targets: commonTargets,
				},
			},
		}, nil
	}

	panic("ModVersionsWithDependencies: " + modID)
}

func (m MockProvider) GetModName(_ context.Context, modReference string) (*ficsit.GetModNameResponse, error) {
	switch modReference {
	case "RefinedPower":
		return &ficsit.GetModNameResponse{Mod: ficsit.GetModNameMod{
			Id:            "DGiLzB3ZErWu2V",
			Mod_reference: "RefinedPower",
			Name:          "Refined Power",
		}}, nil
	case "RefinedRDLib":
		return &ficsit.GetModNameResponse{Mod: ficsit.GetModNameMod{
			Id:            "B24emzbs6xVZQr",
			Mod_reference: "RefinedRDLib",
			Name:          "RefinedRDLib",
		}}, nil
	}

	panic("GetModName: " + modReference)
}

func (m MockProvider) IsOffline() bool {
	return false
}
