package localregistry

import (
	"database/sql"
	"fmt"
)

var migrations = []func(*sql.Tx) error{
	initialSetup,
	addRequiredOnRemote,
}

func applyMigrations(db *sql.DB) error {
	// user_version will store the 1-indexed migration that was last applied
	var nextMigration int
	err := db.QueryRow("PRAGMA user_version;").Scan(&nextMigration)
	if err != nil {
		return fmt.Errorf("failed to get user_version: %w", err)
	}

	for i := nextMigration; i < len(migrations); i++ {
		err := applyMigration(db, i)
		if err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", i, err)
		}
	}

	return nil
}

func applyMigration(db *sql.DB, migrationIndex int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	// Will noop if the transaction was committed
	defer tx.Rollback() //nolint:errcheck

	err = migrations[migrationIndex](tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d;", migrationIndex+1))
	if err != nil {
		return fmt.Errorf("failed to set user_version: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func initialSetup(tx *sql.Tx) error {
	// Create the initial user
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS "versions" (
		    "id" TEXT NOT NULL PRIMARY KEY,
			"mod_reference"	TEXT NOT NULL,
			"version"	TEXT NOT NULL,
			"game_version"	TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS "mod_reference" ON "versions" ("mod_reference");
		CREATE UNIQUE INDEX IF NOT EXISTS "mod_version" ON "versions" ("mod_reference", "version");
		
		CREATE TABLE IF NOT EXISTS "dependencies" (
		    "version_id" TEXT NOT NULL,
		    "dependency" TEXT NOT NULL,
		    "condition" TEXT NOT NULL,
		    "optional" INT NOT NULL,
		    FOREIGN KEY ("version_id") REFERENCES "versions" ("id") ON DELETE CASCADE,
		    PRIMARY KEY ("version_id", "dependency")
		);

		CREATE TABLE IF NOT EXISTS "targets" (
		    "version_id" TEXT NOT NULL,
		    "target_name" TEXT NOT NULL,
		    "link" TEXT NOT NULL,
		    "hash" TEXT NOT NULL,
		    "size" INT NOT NULL,
		    FOREIGN KEY ("version_id") REFERENCES "versions" ("id") ON DELETE CASCADE,
		    PRIMARY KEY ("version_id", "target_name")
	 	);
	`)

	if err != nil {
		return fmt.Errorf("failed to create initial tables: %w", err)
	}

	return nil
}

func addRequiredOnRemote(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE "versions" ADD COLUMN "required_on_remote" INT NOT NULL DEFAULT 1;
	`)

	if err != nil {
		return fmt.Errorf("failed to add required_on_remote column: %w", err)
	}

	return nil
}
