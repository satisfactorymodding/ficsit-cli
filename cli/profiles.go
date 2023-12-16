package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	resolver "github.com/satisfactorymodding/ficsit-resolver"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/utils"
)

const DefaultProfileName = "Default"

var defaultProfile = Profile{
	Name: DefaultProfileName,
}

type ProfilesVersion int

const (
	InitialProfilesVersion = ProfilesVersion(iota)

	// Always last
	nextProfilesVersion
)

type smmProfileFile struct {
	Items []struct {
		ID      string `json:"id"`
		Enabled bool   `json:"enabled"`
	} `json:"items"`
}

type Profiles struct {
	Profiles        map[string]*Profile `json:"profiles"`
	SelectedProfile string              `json:"selected_profile"`
	Version         ProfilesVersion     `json:"version"`
}

type Profile struct {
	Mods            map[string]ProfileMod `json:"mods"`
	Name            string                `json:"name"`
	RequiredTargets []resolver.TargetName `json:"required_targets"`
}

type ProfileMod struct {
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

func InitProfiles() (*Profiles, error) {
	localDir := viper.GetString("local-dir")

	profilesFile := filepath.Join(localDir, viper.GetString("profiles-file"))
	_, err := os.Stat(profilesFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to stat profiles file")
		}

		_, err := os.Stat(localDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrap(err, "failed to read cache directory")
			}

			err = os.MkdirAll(localDir, 0o755)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create cache directory")
			}
		}

		profiles := map[string]*Profile{
			DefaultProfileName: &defaultProfile,
		}

		// Import profiles from SMM if already exists
		smmProfilesDir := filepath.Join(viper.GetString("base-local-dir"), "SatisfactoryModManager", "profiles")
		_, err = os.Stat(smmProfilesDir)
		if err == nil {
			dir, err := os.ReadDir(smmProfilesDir)
			if err == nil {
				for _, entry := range dir {
					if entry.IsDir() {
						manifestFile := filepath.Join(smmProfilesDir, entry.Name(), "manifest.json")
						_, err := os.Stat(manifestFile)
						if err == nil {
							manifestBytes, err := os.ReadFile(manifestFile)
							if err != nil {
								log.Err(err).Str("file", manifestFile).Msg("Failed to read file, not importing profile")
								continue
							}

							var smmProfile smmProfileFile
							if err := json.Unmarshal(manifestBytes, &smmProfile); err != nil {
								log.Err(err).Str("file", manifestFile).Msg("Failed to parse file, not importing profile")
								continue
							}

							profile := &Profile{
								Name: entry.Name(),
								Mods: make(map[string]ProfileMod),
							}

							for _, item := range smmProfile.Items {
								// Explicitly ignore bootstrapper
								if strings.ToLower(item.ID) == "bootstrapper" {
									continue
								}

								profile.Mods[item.ID] = ProfileMod{
									Version: ">=0.0.0",
									Enabled: item.Enabled,
								}
							}

							profiles[entry.Name()] = profile
						}
					}
				}
			}
		}

		bootstrapProfiles := &Profiles{
			Version:  nextProfilesVersion - 1,
			Profiles: profiles,
		}

		if err := bootstrapProfiles.Save(); err != nil {
			return nil, errors.Wrap(err, "failed to save empty profiles")
		}
	}

	profilesData, err := os.ReadFile(profilesFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read profiles")
	}

	var profiles Profiles
	if err := json.Unmarshal(profilesData, &profiles); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal profiles")
	}

	if profiles.Version >= nextProfilesVersion {
		return nil, fmt.Errorf("unknown profiles version: %d", profiles.Version)
	}

	if len(profiles.Profiles) == 0 {
		profiles.Profiles = map[string]*Profile{
			DefaultProfileName: &defaultProfile,
		}
		profiles.SelectedProfile = DefaultProfileName
	}

	if profiles.SelectedProfile == "" || profiles.Profiles[profiles.SelectedProfile] == nil {
		profiles.SelectedProfile = DefaultProfileName
	}

	return &profiles, nil
}

// Save the profiles to the profiles file.
func (p *Profiles) Save() error {
	if viper.GetBool("dry-run") {
		log.Info().Msg("dry-run: skipping profile saving")
		return nil
	}

	profilesFile := filepath.Join(viper.GetString("local-dir"), viper.GetString("profiles-file"))

	log.Info().Str("path", profilesFile).Msg("saving profiles")

	profilesJSON, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal profiles")
	}

	if err := os.WriteFile(profilesFile, profilesJSON, 0o755); err != nil {
		return errors.Wrap(err, "failed to write profiles")
	}

	return nil
}

// AddProfile adds a new profile with the given name to the profiles list.
func (p *Profiles) AddProfile(name string) (*Profile, error) {
	if _, ok := p.Profiles[name]; ok {
		return nil, fmt.Errorf("profile with name %s already exists", name)
	}

	p.Profiles[name] = &Profile{
		Name: name,
	}

	return p.Profiles[name], nil
}

// DeleteProfile deletes the profile with the given name.
func (p *Profiles) DeleteProfile(name string) error {
	if _, ok := p.Profiles[name]; ok {
		delete(p.Profiles, name)

		if p.SelectedProfile == name {
			p.SelectedProfile = DefaultProfileName
		}

		return nil
	}

	return fmt.Errorf("profile with name %s does not exist", name)
}

// GetProfile returns the profile with the given name or nil if it doesn't exist.
func (p *Profiles) GetProfile(name string) *Profile {
	return p.Profiles[name]
}

func (p *Profiles) RenameProfile(ctx *GlobalContext, oldName string, newName string) error {
	if _, ok := p.Profiles[newName]; ok {
		return fmt.Errorf("profile with name %s already exists", newName)
	}

	if _, ok := p.Profiles[oldName]; !ok {
		return fmt.Errorf("profile with name %s does not exist", oldName)
	}

	p.Profiles[oldName].Name = newName
	p.Profiles[newName] = p.Profiles[oldName]
	delete(p.Profiles, oldName)

	if p.SelectedProfile == oldName {
		p.SelectedProfile = newName
	}

	for _, installation := range ctx.Installations.Installations {
		if installation.Profile == oldName {
			installation.Profile = newName
		}
	}

	return nil
}

// AddMod adds a mod to the profile with given version.
func (p *Profile) AddMod(reference string, version string) error {
	if p.Mods == nil {
		p.Mods = make(map[string]ProfileMod)
	}

	if !utils.SemVerRegex.MatchString(version) {
		return errors.New("invalid semver version")
	}

	p.Mods[reference] = ProfileMod{
		Version: version,
		Enabled: true,
	}

	return nil
}

// RemoveMod removes a mod from the profile.
func (p *Profile) RemoveMod(reference string) {
	if p.Mods == nil {
		return
	}

	delete(p.Mods, reference)
}

// HasMod returns true if the profile has a mod with the given reference.
func (p *Profile) HasMod(reference string) bool {
	if p.Mods == nil {
		return false
	}

	_, ok := p.Mods[reference]

	return ok
}

// Resolve resolves all mods and their dependencies.
//
// An optional lockfile can be passed if one exists.
//
// Returns an error if resolution is impossible.
func (p *Profile) Resolve(resolver resolver.DependencyResolver, lockFile *resolver.LockFile, gameVersion int) (*resolver.LockFile, error) {
	toResolve := make(map[string]string)
	for modReference, mod := range p.Mods {
		if mod.Enabled {
			toResolve[modReference] = mod.Version
		}
	}

	resultLockfile, err := resolver.ResolveModDependencies(context.TODO(), toResolve, lockFile, gameVersion, p.RequiredTargets)
	if err != nil {
		return nil, errors.Wrap(err, "failed resolving profile dependencies")
	}

	return resultLockfile, nil
}

func (p *Profile) IsModEnabled(reference string) bool {
	if p.Mods == nil {
		return false
	}

	if mod, ok := p.Mods[reference]; ok {
		return mod.Enabled
	}

	return false
}

func (p *Profile) SetModEnabled(reference string, enabled bool) {
	if p.Mods == nil {
		return
	}

	if _, ok := p.Mods[reference]; !ok {
		return
	}

	p.Mods[reference] = ProfileMod{
		Version: p.Mods[reference].Version,
		Enabled: enabled,
	}
}
