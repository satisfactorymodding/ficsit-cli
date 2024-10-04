package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	resolver "github.com/satisfactorymodding/ficsit-resolver"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/satisfactorymodding/ficsit-cli/cli/cache"
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
			return nil, fmt.Errorf("failed to stat installations file: %w", err)
		}

		_, err := os.Stat(localDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read cache directory: %w", err)
			}

			err = os.MkdirAll(localDir, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create cache directory: %w", err)
			}
		}

		emptyInstallations := Installations{
			Version: nextInstallationsVersion - 1,
		}

		if err := emptyInstallations.Save(); err != nil {
			return nil, fmt.Errorf("failed to save empty installations: %w", err)
		}
	}

	installationsData, err := os.ReadFile(installationsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read installations: %w", err)
	}

	var installations Installations
	if err := json.Unmarshal(installationsData, &installations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal installations: %w", err)
	}

	if installations.Version >= nextInstallationsVersion {
		return nil, fmt.Errorf("unknown installations version: %d", installations.Version)
	}

	return &installations, nil
}

func (i *Installations) Save() error {
	if viper.GetBool("dry-run") {
		slog.Info("dry-run: skipping installation saving")
		return nil
	}

	installationsFile := filepath.Join(viper.GetString("local-dir"), viper.GetString("installations-file"))

	slog.Info("saving installations", slog.String("path", installationsFile))

	installationsJSON, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal installations: %w", err)
	}

	if err := os.WriteFile(installationsFile, installationsJSON, 0o755); err != nil {
		return fmt.Errorf("failed to write installations: %w", err)
	}

	return nil
}

func (i *Installations) AddInstallation(ctx *GlobalContext, installPath string, profile string) (*Installation, error) {
	parsed, err := url.Parse(installPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %w", err)
	}

	absolutePath := installPath
	if parsed.Scheme != "ftp" && parsed.Scheme != "sftp" {
		absolutePath, err = filepath.Abs(installPath)

		if err != nil {
			return nil, fmt.Errorf("could not resolve absolute path of: %s: %w", installPath, err)
		}
	}

	installation := &Installation{
		Path:    absolutePath,
		Profile: profile,
		Vanilla: false,
	}

	if err := installation.Validate(ctx); err != nil {
		return nil, fmt.Errorf("failed to validate installation: %w", err)
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

var rootExecutables = []string{"FactoryGame.exe", "FactoryServer.sh", "FactoryServer.exe", "FactoryGameSteam.exe", "FactoryGameEGS.exe"}

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

	var checkWait errgroup.Group

	for _, executable := range rootExecutables {
		e := executable
		checkWait.Go(func() error {
			exists, err := d.Exists(filepath.Join(i.BasePath(), e))
			if !exists {
				if err != nil {
					return fmt.Errorf("failed reading %s: %w", e, err)
				}
			} else {
				foundExecutable = true
			}
			return nil
		})
	}

	if err = checkWait.Wait(); err != nil {
		return err //nolint:wrapcheck
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

func (i *Installation) lockFilePath(ctx *GlobalContext, platform *Platform) string {
	lockFileName := ctx.Profiles.Profiles[i.Profile].Name
	lockFileName = matchFirstCap.ReplaceAllString(lockFileName, "${1}_${2}")
	lockFileName = matchAllCap.ReplaceAllString(lockFileName, "${1}_${2}")
	lockFileName = lockFileCleaner.ReplaceAllLiteralString(lockFileName, "-")
	lockFileName = strings.ToLower(lockFileName) + "-lock.json"

	return filepath.Join(i.BasePath(), platform.LockfilePath, lockFileName)
}

func (i *Installation) lockfile(ctx *GlobalContext, platform *Platform) (*resolver.LockFile, error) {
	lockfilePath := i.lockFilePath(ctx, platform)

	d, err := i.GetDisk()
	if err != nil {
		return nil, err
	}

	exists, err := d.Exists(lockfilePath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	var lockFile *resolver.LockFile
	lockFileJSON, err := d.Read(lockfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading lockfile: %w", err)
	}

	if err := json.Unmarshal(lockFileJSON, &lockFile); err != nil {
		return nil, fmt.Errorf("failed parsing lockfile: %w", err)
	}

	return lockFile, nil
}

func (i *Installation) writeLockFile(ctx *GlobalContext, platform *Platform, lockfile *resolver.LockFile) error {
	lockfilePath := i.lockFilePath(ctx, platform)

	d, err := i.GetDisk()
	if err != nil {
		return err
	}

	lockfileDir := filepath.Dir(lockfilePath)
	if exists, err := d.Exists(lockfileDir); !exists {
		if err != nil {
			return err
		}

		if err := d.MkDir(lockfileDir); err != nil {
			return fmt.Errorf("failed creating lockfile directory: %w", err)
		}
	}

	marshaledLockfile, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize lockfile json: %w", err)
	}

	if err := d.Write(lockfilePath, marshaledLockfile); err != nil {
		return fmt.Errorf("failed writing lockfile: %w", err)
	}

	return nil
}

func (i *Installation) Wipe() error {
	slog.Info("wiping installation", slog.String("path", i.Path))

	d, err := i.GetDisk()
	if err != nil {
		return err
	}

	modsDirectory := filepath.Join(i.BasePath(), "FactoryGame", "Mods")
	if err := d.Remove(modsDirectory); err != nil {
		return fmt.Errorf("failed removing Mods directory: %w", err)
	}

	return nil
}

func (i *Installation) resolveProfile(ctx *GlobalContext, platform *Platform) (*resolver.LockFile, error) {
	lockFile, err := i.lockfile(ctx, platform)
	if err != nil {
		return nil, err
	}

	depResolver := resolver.NewDependencyResolver(ctx.Provider)

	gameVersion, err := i.getGameVersion(platform)
	if err != nil {
		return nil, fmt.Errorf("failed to detect game version: %w", err)
	}

	lockfile, err := ctx.Profiles.Profiles[i.Profile].Resolve(depResolver, lockFile, gameVersion)
	if err != nil {
		return nil, fmt.Errorf("could not resolve mods: %w", err)
	}

	if err := i.writeLockFile(ctx, platform, lockfile); err != nil {
		return nil, fmt.Errorf("failed to write lockfile: %w", err)
	}

	return lockfile, nil
}

func (i *Installation) GetGameVersion(ctx *GlobalContext) (int, error) {
	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return 0, err
	}
	return i.getGameVersion(platform)
}

func (i *Installation) LockFile(ctx *GlobalContext) (*resolver.LockFile, error) {
	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return nil, err
	}
	return i.lockfile(ctx, platform)
}

func (i *Installation) WriteLockFile(ctx *GlobalContext, lockfile *resolver.LockFile) error {
	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return err
	}
	return i.writeLockFile(ctx, platform, lockfile)
}

type InstallUpdateType string

var (
	InstallUpdateTypeOverall     InstallUpdateType = "overall"
	InstallUpdateTypeModDownload InstallUpdateType = "download"
	InstallUpdateTypeModExtract  InstallUpdateType = "extract"
	InstallUpdateTypeModComplete InstallUpdateType = "complete"
)

type InstallUpdate struct {
	Type     InstallUpdateType
	Item     InstallUpdateItem
	Progress utils.GenericProgress
}

type InstallUpdateItem struct {
	Mod     string
	Version string
}

func (i *Installation) Install(ctx *GlobalContext, updates chan<- InstallUpdate) error {
	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	lockfile := resolver.NewLockfile()

	if !i.Vanilla {
		var err error
		lockfile, err = i.resolveProfile(ctx, platform)
		if err != nil {
			return fmt.Errorf("failed to resolve lockfile: %w", err)
		}
	}

	d, err := i.GetDisk()
	if err != nil {
		return err
	}

	modsDirectory := filepath.Join(i.BasePath(), "FactoryGame", "Mods")
	if err := d.MkDir(modsDirectory); err != nil {
		return fmt.Errorf("failed creating Mods directory: %w", err)
	}

	dir, err := d.ReadDir(modsDirectory)
	if err != nil {
		return fmt.Errorf("failed to read mods directory: %w", err)
	}

	var deleteWait errgroup.Group
	for _, entry := range dir {
		if entry.IsDir() {
			modName := entry.Name()
			mod, hasMod := lockfile.Mods[modName]
			if hasMod {
				_, hasTarget := mod.Targets[platform.TargetName]
				hasMod = hasTarget
			}
			if !hasMod {
				modName := entry.Name()
				modDir := filepath.Join(modsDirectory, modName)
				deleteWait.Go(func() error {
					exists, err := d.Exists(filepath.Join(modDir, ".smm"))
					if err != nil {
						return err
					}

					if exists {
						slog.Info("deleting mod", slog.String("mod_reference", modName))
						if err := d.Remove(modDir); err != nil {
							return fmt.Errorf("failed to delete mod directory: %w", err)
						}
					}

					return nil
				})
			}
		}
	}

	if err := deleteWait.Wait(); err != nil {
		return fmt.Errorf("failed to remove old mods: %w", err)
	}

	slog.Info("starting installation", slog.Int("concurrency", viper.GetInt("concurrent-downloads")), slog.String("path", i.Path))

	errg := errgroup.Group{}
	channelUsers := sync.WaitGroup{}
	downloadSemaphore := make(chan int, viper.GetInt("concurrent-downloads"))
	defer close(downloadSemaphore)

	var modComplete chan int
	if updates != nil {
		channelUsers.Add(1)
		modComplete = make(chan int)
		defer close(modComplete)
		go func() {
			defer channelUsers.Done()
			completed := 0
			for range modComplete {
				completed++
				overallUpdate := InstallUpdate{
					Type: InstallUpdateTypeOverall,
					Progress: utils.GenericProgress{
						Completed: int64(completed),
						Total:     int64(len(lockfile.Mods)),
					},
				}
				updates <- overallUpdate
			}
		}()
	}

	for modReference, version := range lockfile.Mods {
		channelUsers.Add(1)
		modReference := modReference
		version := version
		errg.Go(func() error {
			defer channelUsers.Done()

			target, ok := version.Targets[platform.TargetName]
			if !ok {
				// The resolver validates that the resulting lockfile mods can be installed on the sides where they are required
				// so if the mod is missing this target, it means it is not required on this target
				slog.Info("skipping mod not available for target", slog.String("mod_reference", modReference), slog.String("version", version.Version), slog.String("target", platform.TargetName))
				return nil
			}

			// Only install if a link is provided, otherwise assume mod is already installed
			if target.Link != "" {
				err := downloadAndExtractMod(modReference, version.Version, target.Link, target.Hash, platform.TargetName, modsDirectory, updates, downloadSemaphore, d)
				if err != nil {
					return fmt.Errorf("failed to install %s@%s: %w", modReference, version.Version, err)
				}
			}

			if modComplete != nil {
				modComplete <- 1
			}
			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		return fmt.Errorf("failed to install mods: %w", err)
	}

	if updates != nil {
		if i.Vanilla {
			updates <- InstallUpdate{
				Type: InstallUpdateTypeOverall,
				Progress: utils.GenericProgress{
					Completed: 1,
					Total:     1,
				},
			}
		}

		go func() {
			channelUsers.Wait()
			close(updates)
		}()
	}

	slog.Info("installation completed", slog.String("path", i.Path))

	return nil
}

func (i *Installation) UpdateMods(ctx *GlobalContext, mods []string) error {
	platform, err := i.GetPlatform(ctx)
	if err != nil {
		return err
	}

	lockFile, err := i.lockfile(ctx, platform)
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	resolver := resolver.NewDependencyResolver(ctx.Provider)

	gameVersion, err := i.getGameVersion(platform)
	if err != nil {
		return fmt.Errorf("failed to detect game version: %w", err)
	}

	profile := ctx.Profiles.GetProfile(i.Profile)
	if profile == nil {
		return errors.New("could not find profile " + i.Profile)
	}

	for _, modReference := range mods {
		lockFile = lockFile.Remove(modReference)
	}

	newLockFile, err := profile.Resolve(resolver, lockFile, gameVersion)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	if err := i.writeLockFile(ctx, platform, newLockFile); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

func downloadAndExtractMod(modReference string, version string, link string, hash string, target string, modsDirectory string, updates chan<- InstallUpdate, downloadSemaphore chan int, d disk.Disk) error {
	var downloadUpdates chan utils.GenericProgress

	var wg sync.WaitGroup
	if updates != nil {
		// Forward the inner updates as InstallUpdates
		downloadUpdates = make(chan utils.GenericProgress)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for up := range downloadUpdates {
				updates <- InstallUpdate{
					Item: InstallUpdateItem{
						Mod:     modReference,
						Version: version,
					},
					Type:     InstallUpdateTypeModDownload,
					Progress: up,
				}
			}
		}()
	}

	slog.Info("downloading mod", slog.String("mod_reference", modReference), slog.String("version", version), slog.String("link", link))
	reader, size, err := cache.DownloadOrCache(modReference+"_"+version+"_"+target+".zip", hash, link, downloadUpdates, downloadSemaphore)
	if err != nil {
		return fmt.Errorf("failed to download %s from: %s: %w", modReference, link, err)
	}

	defer reader.Close()

	var extractUpdates chan utils.GenericProgress

	if updates != nil {
		// Forward the inner updates as InstallUpdates
		extractUpdates = make(chan utils.GenericProgress)

		wg.Add(1)
		go func() {
			defer wg.Done()
			for up := range extractUpdates {
				updates <- InstallUpdate{
					Item: InstallUpdateItem{
						Mod:     modReference,
						Version: version,
					},
					Type:     InstallUpdateTypeModExtract,
					Progress: up,
				}
			}
		}()
	}

	slog.Info("extracting mod", slog.String("mod_reference", modReference), slog.String("version", version), slog.String("link", link))
	if err := utils.ExtractMod(reader, size, filepath.Join(modsDirectory, modReference), hash, extractUpdates, d); err != nil {
		return fmt.Errorf("could not extract %s: %w", modReference, err)
	}

	if updates != nil {
		close(downloadUpdates)
		close(extractUpdates)

		updates <- InstallUpdate{
			Type: InstallUpdateTypeModComplete,
			Item: InstallUpdateItem{
				Mod:     modReference,
				Version: version,
			},
		}
	}

	wg.Wait()

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

func (i *Installation) getGameVersion(platform *Platform) (int, error) {
	d, err := i.GetDisk()
	if err != nil {
		return 0, err
	}

	fullPath := filepath.Join(i.BasePath(), platform.VersionPath)

	file, err := d.Read(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed reading version file: %w", err)
	}

	var versionData gameVersionFile
	if err := json.Unmarshal(file, &versionData); err != nil {
		return 0, fmt.Errorf("failed to parse version file json: %w", err)
	}

	return versionData.Changelist, nil
}

func (i *Installation) GetPlatform(ctx *GlobalContext) (*Platform, error) {
	if err := i.Validate(ctx); err != nil {
		return nil, fmt.Errorf("failed to validate installation: %w", err)
	}

	d, err := i.GetDisk()
	if err != nil {
		return nil, err
	}

	for _, platform := range platforms {
		fullPath := filepath.Join(i.BasePath(), platform.VersionPath)
		exists, err := d.Exists(fullPath)
		if !exists {
			if err != nil {
				return nil, fmt.Errorf("failed detecting version file: %w", err)
			}
			continue
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
