package localregistry

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"

	// sqlite driver
	_ "modernc.org/sqlite"
)

var db *sql.DB
var dbWriteMutex = sync.Mutex{}

func Init() error {
	dbPath := filepath.Join(viper.GetString("cache-dir"), "registry.db")

	err := os.MkdirAll(filepath.Dir(dbPath), 0o777)
	if err != nil {
		return fmt.Errorf("failed to create local registry directory: %w", err)
	}

	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set pragmas here because modernc.org/sqlite does not support them in the connection string
	_, err = db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA foreign_keys = ON;
		PRAGMA busy_timeout = 5000;
	`)
	if err != nil {
		return fmt.Errorf("failed to setup connection pragmas: %w", err)
	}

	err = applyMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func Add(modReference string, modVersions []ficsit.ModVersion) {
	dbWriteMutex.Lock()
	defer dbWriteMutex.Unlock()

	tx, err := db.Begin()
	if err != nil {
		slog.Error("failed to start local registry transaction", slog.Any("err", err))
		return
	}
	// In case the transaction is not committed, revert and release
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec("DELETE FROM versions WHERE mod_reference = ?", modReference)
	if err != nil {
		slog.Error("failed to delete existing mod versions from local registry", slog.Any("err", err))
		return
	}

	for _, modVersion := range modVersions {
		l := slog.With(slog.String("mod", modReference), slog.String("version", modVersion.Version))

		_, err = tx.Exec("INSERT INTO versions (id, mod_reference, version, game_version, required_on_remote) VALUES (?, ?, ?, ?, ?)", modVersion.ID, modReference, modVersion.Version, modVersion.GameVersion, modVersion.RequiredOnRemote)
		if err != nil {
			l.Error("failed to insert mod version into local registry", slog.Any("err", err))
			return
		}
		for _, dependency := range modVersion.Dependencies {
			_, err = tx.Exec("INSERT INTO dependencies (version_id, dependency, condition, optional) VALUES (?, ?, ?, ?)", modVersion.ID, dependency.ModID, dependency.Condition, dependency.Optional)
			if err != nil {
				l.Error("failed to insert dependency into local registry", slog.String("dependency", dependency.ModID), slog.Any("err", err))
				return
			}
		}
		for _, target := range modVersion.Targets {
			_, err = tx.Exec("INSERT INTO targets (version_id, target_name, link, hash, size) VALUES (?, ?, ?, ?, ?)", modVersion.ID, target.TargetName, target.Link, target.Hash, target.Size)
			if err != nil {
				l.Error("failed to insert target into local registry", slog.Any("target", target.TargetName), slog.Any("err", err))
				return
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		slog.Error("failed to commit local registry transaction", slog.Any("err", err))
		return
	}
}

func GetModVersions(modReference string) ([]ficsit.ModVersion, error) {
	versionRows, err := db.Query("SELECT id, version, game_version, required_on_remote FROM versions WHERE mod_reference = ?", modReference)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mod versions from local registry: %w", err)
	}
	defer versionRows.Close()

	var versions []ficsit.ModVersion
	for versionRows.Next() {
		var version ficsit.ModVersion
		err = versionRows.Scan(&version.ID, &version.Version, &version.GameVersion, &version.RequiredOnRemote)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version row: %w", err)
		}

		dependencies, err := getVersionDependencies(version.ID)
		if err != nil {
			return nil, err
		}

		version.Dependencies = dependencies

		targets, err := getVersionTargets(version.ID)
		if err != nil {
			return nil, err
		}

		version.Targets = targets

		versions = append(versions, version)
	}

	return versions, nil
}

func getVersionDependencies(versionID string) ([]ficsit.Dependency, error) {
	var dependencies []ficsit.Dependency
	dependencyRows, err := db.Query("SELECT dependency, condition, optional FROM dependencies WHERE version_id = ?", versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dependencies from local registry: %w", err)
	}
	defer dependencyRows.Close()

	for dependencyRows.Next() {
		var dependency ficsit.Dependency
		err = dependencyRows.Scan(&dependency.ModID, &dependency.Condition, &dependency.Optional)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dependency row: %w", err)
		}
		dependencies = append(dependencies, dependency)
	}

	return dependencies, nil
}

func getVersionTargets(versionID string) ([]ficsit.Target, error) {
	var targets []ficsit.Target
	targetRows, err := db.Query("SELECT target_name, link, hash, size FROM targets WHERE version_id = ?", versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch targets from local registry: %w", err)
	}
	defer targetRows.Close()

	for targetRows.Next() {
		var target ficsit.Target
		err = targetRows.Scan(&target.TargetName, &target.Link, &target.Hash, &target.Size)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target row: %w", err)
		}
		targets = append(targets, target)
	}

	return targets, nil
}
