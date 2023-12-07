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

var commonTargets = []ficsit.ModVersionsWithDependenciesModVersionsVersionTargetsVersionTarget{
	{
		TargetName: ficsit.TargetNameWindows,
		Link:       "/v1/version/7QcfNdo5QAAyoC/Windows/download",
		Hash:       "62f5c84eca8480b3ffe7d6c90f759e3b463f482530e27d854fd48624fdd3acc9",
	},
	{
		TargetName: ficsit.TargetNameWindowsserver,
		Link:       "/v1/version/7QcfNdo5QAAyoC/WindowsServer/download",
		Hash:       "8a83fcd4abece4192038769cc672fff6764d72c32fb6c7a8c58d66156bb07917",
	},
	{
		TargetName: ficsit.TargetNameLinuxserver,
		Link:       "/v1/version/7QcfNdo5QAAyoC/LinuxServer/download",
		Hash:       "8739c76e681f900923b900c9df0ef75cf421d39cabb54650c4b9ad19b6a76d85",
	},
}

func (m MockProvider) ModVersionsWithDependencies(_ context.Context, modID string) (*ficsit.ModVersionsWithDependenciesResponse, error) {
	switch modID {
	case "RefinedPower":
		return &ficsit.ModVersionsWithDependenciesResponse{
			Mod: ficsit.ModVersionsWithDependenciesMod{
				Id: "DGiLzB3ZErWu2V",
				Versions: []ficsit.ModVersionsWithDependenciesModVersionsVersion{
					{
						Id:      "Eqgr4VcB8y1z9a",
						Version: "3.2.13",
						Link:    "/v1/version/Eqgr4VcB8y1z9a/download",
						Hash:    "8cabf9245e3f2a01b95cd3d39d98e407cfeccf355c19f1538fcbf868f81de008",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "ModularUI",
								Condition: "^2.1.11",
								Optional:  false,
							},
							{
								Mod_id:    "RefinedRDLib",
								Condition: "^1.1.7",
								Optional:  false,
							},
							{
								Mod_id:    "SML",
								Condition: "^3.6.1",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "BwVKMJNP8doDLg",
						Version: "3.2.11",
						Link:    "/v1/version/BwVKMJNP8doDLg/download",
						Hash:    "b64aa7b3a4766295323eac47d432e0d857d042c9cfb1afdd16330483b0476c89",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "ModularUI",
								Condition: "^2.1.10",
								Optional:  false,
							},
							{
								Mod_id:    "RefinedRDLib",
								Condition: "^1.1.6",
								Optional:  false,
							},
							{
								Mod_id:    "SML",
								Condition: "^3.6.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "4XTjMpqFngbu9r",
						Version: "3.2.10",
						Link:    "/v1/version/4XTjMpqFngbu9r/download",
						Hash:    "093f92c6d52c853bade386d5bc79cf103b27fb6e9d6f806850929b866ff98222",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "ModularUI",
								Condition: "^2.1.9",
								Optional:  false,
							},
							{
								Mod_id:    "RefinedRDLib",
								Condition: "^1.1.5",
								Optional:  false,
							},
							{
								Mod_id:    "SML",
								Condition: "^3.6.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
				},
			},
		}, nil
	case "AreaActions":
		return &ficsit.ModVersionsWithDependenciesResponse{
			Mod: ficsit.ModVersionsWithDependenciesMod{
				Id: "6vQ6ckVYFiidDh",
				Versions: []ficsit.ModVersionsWithDependenciesModVersionsVersion{
					{
						Id:      "5KMXBkdAz5YJe",
						Version: "1.6.7",
						Link:    "/v1/version/5KMXBkdAz5YJe/download",
						Hash:    "0baa673eea245b8ec5fe203a70b98deb666d85e27fb6ce9201e3c0fa3aaedcbe",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.4.1",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "EtEbwJj3smMn3o",
						Version: "1.6.6",
						Link:    "/v1/version/EtEbwJj3smMn3o/download",
						Hash:    "b64aa7b3a4766295323eac47d432e0d857d042c9cfb1afdd16330483b0476c89",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.2.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "9uw1eDwgrQs279",
						Version: "1.6.5",
						Link:    "/v1/version/9uw1eDwgrQs279/download",
						Hash:    "427a93383fe8a8557096666b7e81bf5fb25f54a5428248904f52adc4dc34d60c",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.0.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
				},
			},
		}, nil
	case "RefinedRDLib":
		return &ficsit.ModVersionsWithDependenciesResponse{
			Mod: ficsit.ModVersionsWithDependenciesMod{
				Id: "B24emzbs6xVZQr",
				Versions: []ficsit.ModVersionsWithDependenciesModVersionsVersion{
					{
						Id:      "2XcE6RUzGhZW7p",
						Version: "1.1.7",
						Link:    "/v1/version/2XcE6RUzGhZW7p/download",
						Hash:    "034f3a7862d0153768e1a95d29d47a9d08ebcb7ff0fc8f9f2cb59147b09f16dd",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.6.1",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "52RMLEigqT5Ksn",
						Version: "1.1.6",
						Link:    "/v1/version/52RMLEigqT5Ksn/download",
						Hash:    "9577e401e1a12a29657c8e3ed0cff34815009504dc62fc1a335b1e7a3b6fed12",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.6.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "F4HY9eP4D5XjWQ",
						Version: "1.1.5",
						Link:    "/v1/version/F4HY9eP4D5XjWQ/download",
						Hash:    "9cbeae078e28a661ebe15642e6d8f652c6c40c50dabd79a0781e25b84ed9bddf",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.6.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
				},
			},
		}, nil
	case "ModularUI":
		return &ficsit.ModVersionsWithDependenciesResponse{
			Mod: ficsit.ModVersionsWithDependenciesMod{
				Id: "As2uJmQLLxjXLG",
				Versions: []ficsit.ModVersionsWithDependenciesModVersionsVersion{
					{
						Id:      "7ay11W9MAv6MHs",
						Version: "2.1.12",
						Link:    "/v1/version/7ay11W9MAv6MHs/download",
						Hash:    "a0de64c02448f9e37903e7569cc6ceee67f8e018f2774aac9cf295704b9e4696",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.6.1",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "4YuL9UbCDdzm68",
						Version: "2.1.11",
						Link:    "/v1/version/4YuL9UbCDdzm68/download",
						Hash:    "b70658bfa74c132530046bee886c3c0f0277b95339b4fc67da6207cbd2cd422d",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.6.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "5yY2zmx5nTyhWv",
						Version: "2.1.10",
						Link:    "/v1/version/5yY2zmx5nTyhWv/download",
						Hash:    "7c523c9e6263a0b182ed42fe4d4de40aada10c17b1b344219618cd39055870bd",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.6.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
				},
			},
		}, nil
	case "ThisModDoesNotExist$$$":
		return &ficsit.ModVersionsWithDependenciesResponse{}, nil
	case "FicsitRemoteMonitoring":
		return &ficsit.ModVersionsWithDependenciesResponse{
			Mod: ficsit.ModVersionsWithDependenciesMod{
				Id: "9LguyCdDUrpT9N",
				Versions: []ficsit.ModVersionsWithDependenciesModVersionsVersion{
					{
						Id:      "7ay11W9MAv6MHs",
						Version: "0.10.1",
						Link:    "/v1/version/9LguyCdDUrpT9N/download",
						Hash:    "9278b37653ad33dd859875929b15cd1f8aba88d0ea65879df2db1ae8808029d4",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.6.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "DYvfwan5tYqZKE",
						Version: "0.10.0",
						Link:    "/v1/version/DYvfwan5tYqZKE/download",
						Hash:    "8666b37b24188c3f56b1dad6f1d437c1127280381172a1046e85142e7cb81c64",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.5.0",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
					{
						Id:      "918KMrX94xFpVw",
						Version: "0.9.8",
						Link:    "/v1/version/918KMrX94xFpVw/download",
						Hash:    "d4fed641b6ecb25b9191f4dd7210576e9bd7bc644abcb3ca592200ccfd08fc44",
						Dependencies: []ficsit.ModVersionsWithDependenciesModVersionsVersionDependenciesVersionDependency{
							{
								Mod_id:    "SML",
								Condition: "^3.4.1",
								Optional:  false,
							},
						},
						Targets: commonTargets,
					},
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
