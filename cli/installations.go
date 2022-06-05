package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/satisfactorymodding/ficsit-cli/utils"

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

	installationsFile := filepath.Join(localDir, viper.GetString("installations-file"))
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

	installationsFile := filepath.Join(viper.GetString("local-dir"), viper.GetString("installations-file"))

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
	absolutePath, err := filepath.Abs(installPath)

	if err != nil {
		return nil, errors.Wrap(err, "could not resolve absolute path of: "+installPath)
	}

	installation := &Installation{
		Path:    absolutePath,
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

	_, err := os.Stat(filepath.Join(i.Path, "FactoryGame.exe"))
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryGame.exe")
		}
	} else {
		foundExecutable = true
	}

	_, err = os.Stat(filepath.Join(i.Path, "FactoryServer.sh"))
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryServer.sh")
		}
	} else {
		foundExecutable = true
	}

	_, err = os.Stat(filepath.Join(i.Path, "FactoryServer.exe"))
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

var (
	lockFileCleaner = regexp.MustCompile(`[^a-zA-Z\d]]`)
	matchFirstCap   = regexp.MustCompile(`(.)([A-Z][a-z]+)`)
	matchAllCap     = regexp.MustCompile(`([a-z\d])([A-Z])`)
)

func (i *Installation) LockFilePath(ctx *GlobalContext) (string, error) {
	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return "", err
	}

	lockFileName := ctx.Profiles.Profiles[i.Profile].Name
	lockFileName = matchFirstCap.ReplaceAllString(lockFileName, "${1}_${2}")
	lockFileName = matchAllCap.ReplaceAllString(lockFileName, "${1}_${2}")
	lockFileName = lockFileCleaner.ReplaceAllLiteralString(lockFileName, "-")
	lockFileName = strings.ToLower(lockFileName) + "-lock.json"

	return filepath.Join(i.Path, platform.LockfilePath, lockFileName), nil
}

func (i *Installation) LockFile(ctx *GlobalContext) (*LockFile, error) {
	lockfilePath, err := i.LockFilePath(ctx)
	if err != nil {
		return nil, err
	}

	var lockFile *LockFile
	lockFileJSON, err := os.ReadFile(lockfilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "failed reading lockfile")
		}
	} else {
		if err := json.Unmarshal(lockFileJSON, &lockFile); err != nil {
			return nil, errors.Wrap(err, "failed parsing lockfile")
		}
	}

	return lockFile, nil
}

func (i *Installation) WriteLockFile(ctx *GlobalContext, lockfile LockFile) error {
	lockfilePath, err := i.LockFilePath(ctx)
	if err != nil {
		return err
	}

	marshaledLockfile, err := json.MarshalIndent(lockfile, "", "  ")

	if err != nil {
		return errors.Wrap(err, "failed to serialize lockfile json")
	}

	if err := os.WriteFile(lockfilePath, marshaledLockfile, 0777); err != nil {
		return errors.Wrap(err, "failed writing lockfile")
	}

	return nil
}

type InstallUpdate struct {
	ModName          string
	OverallProgress  float64
	DownloadProgress float64
	ExtractProgress  float64
}

func (i *Installation) Install(ctx *GlobalContext, updates chan InstallUpdate) error {
	if err := i.Validate(ctx); err != nil {
		return errors.Wrap(err, "failed to validate installation")
	}

	lockFile, err := i.LockFile(ctx)
	if err != nil {
		return err
	}

	resolver := NewDependencyResolver(ctx.APIClient)

	gameVersion, err := i.GetGameVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to detect game version")
	}

	lockfile, err := ctx.Profiles.Profiles[i.Profile].Resolve(resolver, lockFile, gameVersion)

	if err != nil {
		return errors.Wrap(err, "could not resolve mods")
	}

	modsDirectory := filepath.Join(i.Path, "FactoryGame", "Mods")
	if err := os.MkdirAll(modsDirectory, 0777); err != nil {
		return errors.Wrap(err, "failed creating Mods directory")
	}

	dir, err := os.ReadDir(modsDirectory)
	if err != nil {
		return errors.Wrap(err, "failed to read mods directory")
	}

	for _, entry := range dir {
		if entry.IsDir() {
			if _, ok := lockfile[entry.Name()]; !ok {
				if err := os.RemoveAll(filepath.Join(modsDirectory, entry.Name())); err != nil {
					return errors.Wrap(err, "failed to delete mod directory")
				}
			}
		}
	}

	completed := 0
	for modReference, version := range lockfile {
		// Only install if a link is provided, otherwise assume mod is already installed
		if version.Link != "" {
			downloading := true

			var genericUpdates chan utils.GenericUpdate
			if updates != nil {
				genericUpdates = make(chan utils.GenericUpdate)

				go func() {
					update := InstallUpdate{
						ModName:          modReference,
						OverallProgress:  float64(completed) / float64(len(lockfile)),
						DownloadProgress: 0,
						ExtractProgress:  0,
					}

					select {
					case updates <- update:
					default:
					}

					for up := range genericUpdates {
						if downloading {
							update.DownloadProgress = up.Progress
						} else {
							update.DownloadProgress = 1
							update.ExtractProgress = up.Progress
						}

						select {
						case updates <- update:
						default:
						}
					}
				}()
			}

			log.Info().Str("mod_reference", modReference).Str("version", version.Version).Str("link", version.Link).Msg("downloading mod")
			reader, size, err := utils.DownloadOrCache(modReference+"_"+version.Version+".zip", version.Hash, version.Link, genericUpdates)
			if err != nil {
				return errors.Wrap(err, "failed to download "+modReference+" from: "+version.Link)
			}

			downloading = false

			log.Info().Str("mod_reference", modReference).Str("version", version.Version).Str("link", version.Link).Msg("extracting mod")
			if err := utils.ExtractMod(reader, size, filepath.Join(modsDirectory, modReference), genericUpdates); err != nil {
				return errors.Wrap(err, "could not extract "+modReference)
			}

			if genericUpdates != nil {
				close(genericUpdates)
			}
		}

		completed++
	}

	if err := i.WriteLockFile(ctx, lockfile); err != nil {
		return err
	}

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

type gameVersionFile struct {
	MajorVersion         int    `json:"MajorVersion"`
	MinorVersion         int    `json:"MinorVersion"`
	PatchVersion         int    `json:"PatchVersion"`
	Changelist           int    `json:"Changelist"`
	CompatibleChangelist int    `json:"CompatibleChangelist"`
	IsLicenseeVersion    int    `json:"IsLicenseeVersion"`
	IsPromotedBuild      int    `json:"IsPromotedBuild"`
	BranchName           string `json:"BranchName"`
	BuildID              string `json:"BuildId"`
}

func (i *Installation) GetGameVersion(ctx *GlobalContext) (int, error) {
	if err := i.Validate(ctx); err != nil {
		return 0, errors.Wrap(err, "failed to validate installation")
	}

	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return 0, err
	}

	fullPath := filepath.Join(i.Path, platform.VersionPath)
	file, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, errors.Wrap(err, "could not find game version file")
		}
		return 0, errors.Wrap(err, "failed reading version file")
	}

	var versionData gameVersionFile
	if err := json.Unmarshal(file, &versionData); err != nil {
		return 0, errors.Wrap(err, "failed to parse version file json")
	}

	return versionData.Changelist, nil
}

func (i *Installation) GetPlatform(ctx *GlobalContext) (*Platform, error) {
	if err := i.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to validate installation")
	}

	for _, platform := range platforms {
		fullPath := filepath.Join(i.Path, platform.VersionPath)
		_, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return nil, errors.Wrap(err, "failed detecting version file")
			}
		}
		return &platform, nil
	}

	return nil, errors.New("no platform detected")
}
