package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type InstallationsVersion int

const (
	InitialInstallationsVersion = InstallationsVersion(iota)

	// Always last
	nextInstallationsVersion
)

type Installations struct {
	Version              InstallationsVersion `json:"version"`
	Installations        []*Installation      `json:"installations"`
	SelectedInstallation string               `json:"selected_installation"`
}

type Installation struct {
	Path    string `json:"path"`
	Profile string `json:"profile"`
}

func InitInstallations() (*Installations, error) {
	localDir := viper.GetString("local-dir")

	installationsFile := path.Join(localDir, viper.GetString("installations-file"))
	_, err := os.Stat(installationsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to stat installations file")
		}

		_, err := os.Stat(localDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrap(err, "failed to read cache directory")
			}

			err = os.MkdirAll(localDir, 0755)
			if err != nil {
				return nil, errors.Wrap(err, "failed to create cache directory")
			}
		}

		emptyInstallations := Installations{
			Version: nextInstallationsVersion - 1,
		}

		if err := emptyInstallations.Save(); err != nil {
			return nil, errors.Wrap(err, "failed to save empty installations")
		}
	}

	installationsData, err := os.ReadFile(installationsFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read installations")
	}

	var installations Installations
	if err := json.Unmarshal(installationsData, &installations); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal installations")
	}

	if installations.Version >= nextInstallationsVersion {
		return nil, fmt.Errorf("unknown installations version: %d", installations.Version)
	}

	return &installations, nil
}

func (i *Installations) Save() error {
	if viper.GetBool("dry-run") {
		log.Info().Msg("dry-run: skipping installation saving")
		return nil
	}

	installationsFile := path.Join(viper.GetString("local-dir"), viper.GetString("installations-file"))

	log.Info().Str("path", installationsFile).Msg("saving installations")

	installationsJSON, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal installations")
	}

	if err := os.WriteFile(installationsFile, installationsJSON, 0755); err != nil {
		return errors.Wrap(err, "failed to write installations")
	}

	return nil
}

func (i *Installations) AddInstallation(ctx *GlobalContext, installPath string, profile string) (*Installation, error) {
	installation := &Installation{
		Path:    installPath,
		Profile: profile,
	}

	if err := installation.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to validate installation")
	}

	newStat, err := os.Stat(installation.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to stat installation directory")
	}

	found := false
	for _, install := range i.Installations {
		stat, err := os.Stat(install.Path)
		if err != nil {
			continue
		}

		found = os.SameFile(newStat, stat)
		if found {
			break
		}
	}

	if found {
		return nil, errors.New("installation already present")
	}

	i.Installations = append(i.Installations, installation)

	return installation, nil
}

func (i *Installations) GetInstallation(installPath string) *Installation {
	for _, install := range i.Installations {
		if install.Path == installPath {
			return install
		}
	}

	return nil
}

func (i *Installations) DeleteInstallation(installPath string) error {
	found := -1
	for j, install := range i.Installations {
		if install.Path == installPath {
			found = j
			break
		}
	}

	if found == -1 {
		return errors.New("installation not found")
	}

	i.Installations = append(i.Installations[:found], i.Installations[found+1:]...)

	return nil
}

func (i *Installation) Validate(ctx *GlobalContext) error {
	found := false
	for _, p := range ctx.Profiles.Profiles {
		if p.Name == i.Profile {
			found = true
			break
		}
	}

	if !found {
		return errors.New("profile not found")
	}

	foundExecutable := false

	_, err := os.Stat(path.Join(i.Path, "FactoryGame.exe"))
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryGame.exe")
		}
	} else {
		foundExecutable = true
	}

	_, err = os.Stat(path.Join(i.Path, "FactoryServer.sh"))
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryServer.sh")
		}
	} else {
		foundExecutable = true
	}

	_, err = os.Stat(path.Join(i.Path, "FactoryServer.exe"))
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryServer.exe")
		}
	} else {
		foundExecutable = true
	}

	if !foundExecutable {
		return errors.New("did not find game executable in " + i.Path)
	}

	return nil
}

func (i *Installation) Install(ctx *GlobalContext) error {
	if err := i.Validate(ctx); err != nil {
		return errors.Wrap(err, "failed to validate installation")
	}

	// TODO Install from lockfile

	return nil
}

func (i *Installation) SetProfile(ctx *GlobalContext, profile string) error {
	found := false
	for _, p := range ctx.Profiles.Profiles {
		if p.Name == profile {
			found = true
			break
		}
	}

	if !found {
		return errors.New("could not find profile: " + profile)
	}

	i.Profile = profile

	return nil
}
