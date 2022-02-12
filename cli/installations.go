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
	cacheDir := viper.GetString("cache-dir")

	installationsFile := path.Join(cacheDir, viper.GetString("installations-file"))
	_, err := os.Stat(installationsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed to stat installations file")
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

	installationsFile := path.Join(viper.GetString("cache-dir"), viper.GetString("installations-file"))

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
	for _, install := range ctx.Installations.Installations {
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

func (i *Installations) DeleteInstallation(path string) error {
	var idxToDelete = -1
	for i, install := range i.Installations {
		if install.Path == path {
			idxToDelete = i
			break
		}
	}

	if idxToDelete < 0 {
		return fmt.Errorf("installation with path %s does not exist", path)
	} else {
		copy(i.Installations[idxToDelete:], i.Installations[idxToDelete+1:])
		i.Installations = i.Installations[:len(i.Installations)-1]
	}

	return nil
}

func (i *Installations) GetInstallation(installPath string) *Installation {
	for _, install := range i.Installations {
		if install.Path == installPath {
			return install
		}
	}

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

	// TODO Validate installation path

	return nil
}

func (i *Installation) Install(ctx *GlobalContext) error {
	if err := i.Validate(ctx); err != nil {
		return errors.Wrap(err, "failed to validate installation")
	}

	return nil
}
