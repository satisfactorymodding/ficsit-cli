package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/cli/disk"
	"github.com/satisfactorymodding/ficsit-cli/utils"
)

type InstallationsVersion int

const (
	InitialInstallationsVersion = InstallationsVersion(iota)

	// Always last
	nextInstallationsVersion
)

type Installations struct {
	SelectedInstallation string               `json:"selected_installation"`
	Installations        []*Installation      `json:"installations"`
	Version              InstallationsVersion `json:"version"`
}

type Installation struct {
	DiskInstance disk.Disk `json:"-"`
	Path         string    `json:"path"`
	Profile      string    `json:"profile"`
	Vanilla      bool      `json:"vanilla"`
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

			err = os.MkdirAll(localDir, 0o755)
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

	if err := os.WriteFile(installationsFile, installationsJSON, 0o755); err != nil {
		return errors.Wrap(err, "failed to write installations")
	}

	return nil
}

func (i *Installations) AddInstallation(ctx *GlobalContext, installPath string, profile string) (*Installation, error) {
	parsed, err := url.Parse(installPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse path")
	}

	absolutePath := installPath
	if parsed.Scheme != "ftp" && parsed.Scheme != "sftp" {
		absolutePath, err = filepath.Abs(installPath)

		if err != nil {
			return nil, errors.Wrap(err, "could not resolve absolute path of: "+installPath)
		}
	}

	installation := &Installation{
		Path:    absolutePath,
		Profile: profile,
		Vanilla: false,
	}

	if err := installation.Validate(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to validate installation")
	}

	found := false
	for _, install := range i.Installations {
		if filepath.Clean(installation.Path) == filepath.Clean(install.Path) {
			found = true
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

	d, err := i.GetDisk()
	if err != nil {
		return err
	}

	foundExecutable := false

	err = d.Exists(filepath.Join(i.BasePath(), "FactoryGame.exe"))
	if err != nil {
		if !d.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryGame.exe")
		}
	} else {
		foundExecutable = true
	}

	err = d.Exists(filepath.Join(i.BasePath(), "FactoryServer.sh"))
	if err != nil {
		if !d.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryServer.sh")
		}
	} else {
		foundExecutable = true
	}

	err = d.Exists(filepath.Join(i.BasePath(), "FactoryServer.exe"))
	if err != nil {
		if !d.IsNotExist(err) {
			return errors.Wrap(err, "failed reading FactoryServer.exe")
		}
	} else {
		foundExecutable = true
	}

	if !foundExecutable {
		return errors.New("did not find game executable in " + i.BasePath())
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

	return filepath.Join(i.BasePath(), platform.LockfilePath, lockFileName), nil
}

func (i *Installation) LockFile(ctx *GlobalContext) (*LockFile, error) {
	lockfilePath, err := i.LockFilePath(ctx)
	if err != nil {
		return nil, err
	}

	d, err := i.GetDisk()
	if err != nil {
		return nil, err
	}

	var lockFile *LockFile
	lockFileJSON, err := d.Read(lockfilePath)
	if err != nil {
		if !d.IsNotExist(err) {
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

	d, err := i.GetDisk()
	if err != nil {
		return err
	}

	lockfileDir := filepath.Dir(lockfilePath)
	if err := d.Exists(lockfileDir); d.IsNotExist(err) {
		if err := d.MkDir(lockfileDir); err != nil {
			return errors.Wrap(err, "failed creating lockfile directory")
		}
	}

	marshaledLockfile, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to serialize lockfile json")
	}

	if err := d.Write(lockfilePath, marshaledLockfile); err != nil {
		return errors.Wrap(err, "failed writing lockfile")
	}

	return nil
}

func (i *Installation) ResolveProfile(ctx *GlobalContext) (LockFile, error) {
	lockFile, err := i.LockFile(ctx)
	if err != nil {
		return nil, err
	}

	resolver := NewDependencyResolver(ctx.APIClient)

	gameVersion, err := i.GetGameVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to detect game version")
	}

	lockfile, err := ctx.Profiles.Profiles[i.Profile].Resolve(resolver, lockFile, gameVersion)
	if err != nil {
		return nil, errors.Wrap(err, "could not resolve mods")
	}

	if err := i.WriteLockFile(ctx, lockfile); err != nil {
		return nil, errors.Wrap(err, "failed to write lockfile")
	}

	return lockfile, nil
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

	lockfile := make(LockFile)

	if !i.Vanilla {
		var err error
		lockfile, err = i.ResolveProfile(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to resolve lockfile")
		}
	}

	d, err := i.GetDisk()
	if err != nil {
		return err
	}

	modsDirectory := filepath.Join(i.BasePath(), "FactoryGame", "Mods")
	if err := d.MkDir(modsDirectory); err != nil {
		return errors.Wrap(err, "failed creating Mods directory")
	}

	dir, err := d.ReadDir(modsDirectory)
	if err != nil {
		return errors.Wrap(err, "failed to read mods directory")
	}

	for _, entry := range dir {
		if entry.IsDir() {
			if _, ok := lockfile[entry.Name()]; !ok {
				modDir := filepath.Join(modsDirectory, entry.Name())
				err := d.Exists(filepath.Join(modDir, ".smm"))
				if err == nil {
					log.Info().Str("mod_reference", entry.Name()).Msg("deleting mod")
					if err := d.Remove(modDir); err != nil {
						return errors.Wrap(err, "failed to delete mod directory")
					}
				}
			}
		}
	}

	downloading := true
	completed := 0

	var genericUpdates chan utils.GenericUpdate
	if updates != nil {
		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Wait()

		genericUpdates = make(chan utils.GenericUpdate)
		defer close(genericUpdates)

		go func() {
			defer wg.Done()

			update := InstallUpdate{
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

				if up.ModReference != nil {
					update.ModName = *up.ModReference
				}

				update.OverallProgress = float64(completed) / float64(len(lockfile))

				select {
				case updates <- update:
				default:
				}
			}
		}()
	}

	for modReference, version := range lockfile {
		// Only install if a link is provided, otherwise assume mod is already installed
		if version.Link != "" {
			downloading = true

			if genericUpdates != nil {
				genericUpdates <- utils.GenericUpdate{ModReference: &modReference}
			}

			log.Info().Str("mod_reference", modReference).Str("version", version.Version).Str("link", version.Link).Msg("downloading mod")
			reader, size, err := utils.DownloadOrCache(modReference+"_"+version.Version+".zip", version.Hash, version.Link, genericUpdates)
			if err != nil {
				return errors.Wrap(err, "failed to download "+modReference+" from: "+version.Link)
			}

			downloading = false

			log.Info().Str("mod_reference", modReference).Str("version", version.Version).Str("link", version.Link).Msg("extracting mod")
			if err := utils.ExtractMod(reader, size, filepath.Join(modsDirectory, modReference), version.Hash, genericUpdates, d); err != nil {
				return errors.Wrap(err, "could not extract "+modReference)
			}
		}

		completed++
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
	BranchName           string `json:"BranchName"`
	BuildID              string `json:"BuildId"`
	MajorVersion         int    `json:"MajorVersion"`
	MinorVersion         int    `json:"MinorVersion"`
	PatchVersion         int    `json:"PatchVersion"`
	Changelist           int    `json:"Changelist"`
	CompatibleChangelist int    `json:"CompatibleChangelist"`
	IsLicenseeVersion    int    `json:"IsLicenseeVersion"`
	IsPromotedBuild      int    `json:"IsPromotedBuild"`
}

func (i *Installation) GetGameVersion(ctx *GlobalContext) (int, error) {
	if err := i.Validate(ctx); err != nil {
		return 0, errors.Wrap(err, "failed to validate installation")
	}

	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return 0, err
	}

	d, err := i.GetDisk()
	if err != nil {
		return 0, err
	}

	fullPath := filepath.Join(i.BasePath(), platform.VersionPath)
	file, err := d.Read(fullPath)
	if err != nil {
		if d.IsNotExist(err) {
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

	d, err := i.GetDisk()
	if err != nil {
		return nil, err
	}

	for _, platform := range platforms {
		fullPath := filepath.Join(i.BasePath(), platform.VersionPath)
		err := d.Exists(fullPath)
		if err != nil {
			if d.IsNotExist(err) {
				continue
			}
			return nil, errors.Wrap(err, "failed detecting version file")
		}
		return &platform, nil
	}

	return nil, errors.New("no platform detected")
}

func (i *Installation) GetDisk() (disk.Disk, error) {
	if i.DiskInstance != nil {
		return i.DiskInstance, nil
	}

	var err error
	i.DiskInstance, err = disk.FromPath(i.Path)
	return i.DiskInstance, err
}

func (i *Installation) BasePath() string {
	parsed, err := url.Parse(i.Path)
	if err != nil {
		return i.Path
	}

	if parsed.Scheme != "ftp" && parsed.Scheme != "sftp" {
		return i.Path
	}

	return parsed.Path
}
