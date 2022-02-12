package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/utils"
	tea_utils "github.com/satisfactorymodding/ficsit-cli/tea/utils"
	"github.com/spf13/viper"
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

type Profiles struct {
	Version         ProfilesVersion     `json:"version"`
	Profiles        map[string]*Profile `json:"profiles"`
	SelectedProfile string              `json:"selected_profile"`
}

type Profile struct {
	Name string                `json:"name"`
	Mods map[string]ProfileMod `json:"mods"`
}

type ProfileMod struct {
	Name             string `json:"name"`
	ID               string `json:"id"`
	Reference        string `json:"reference"`
	Version          string `json:"version"`
	InstalledVersion string `json:"installed_version"`
}

func InitProfiles() (*Profiles, error) {
	cacheDir := viper.GetString("cache-dir")

	profilesFile := path.Join(cacheDir, viper.GetString("profiles-file"))
	_, err := os.Stat(profilesFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to stat profiles file")
		}

		_, err := os.Stat(cacheDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrap(err, "failed to read cache directory")
			}

			err = os.MkdirAll(cacheDir, 0755)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create cache directory")
			}
		}

		emptyProfiles := Profiles{
			Version: nextProfilesVersion - 1,
			Profiles: map[string]*Profile{
				DefaultProfileName: &defaultProfile,
			},
		}

		if err := emptyProfiles.Save(); err != nil {
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

	profilesFile := path.Join(viper.GetString("cache-dir"), viper.GetString("profiles-file"))

	log.Info().Str("path", profilesFile).Msg("saving profiles")

	profilesJSON, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal profiles")
	}

	if err := os.WriteFile(profilesFile, profilesJSON, 0755); err != nil {
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

func (p *Profiles) RenameProfile(oldName string, newName string) error {
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

	return nil
}

// AddMod adds a mod to the profile with given version.
func (p *Profile) AddMod(mod tea_utils.Mod, version string) error {
	if p.Mods == nil {
		p.Mods = make(map[string]ProfileMod)
	}

	if !utils.SemVerRegex.MatchString(version) {
		return errors.New("invalid semver version")
	}

	p.Mods[mod.Reference] = ProfileMod{
		Name:      mod.Name,
		ID:        mod.ID,
		Reference: mod.Reference,
		Version:   version,
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

// HasMod returns true if the profile has a mod with the given reference
func (p *Profile) HasMod(reference string) bool {
	if p.Mods == nil {
		return false
	}

	_, ok := p.Mods[reference]

	return ok
}
