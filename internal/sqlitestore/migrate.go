package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// migrations[i] upgrades the schema from version i to i+1 (PRAGMA user_version).
var migrations = []string{
	// v1: Relationship becomes unique per (source, target, type) instead of per pair.
	// The first CREATE matches the legacy schema, so databases created by the old code are upgraded in place.
	`
	CREATE TABLE IF NOT EXISTS Person (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		occupation TEXT,
		image_url TEXT,
		twitter TEXT,
		description TEXT
	);
	CREATE TABLE IF NOT EXISTS User (
		id TEXT PRIMARY KEY,
		login TEXT UNIQUE,
		email TEXT UNIQUE,
		password TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS Session (
		id TEXT PRIMARY KEY,
		user_id TEXT,
		expires_at INTEGER,
		FOREIGN KEY (user_id) REFERENCES User(id)
	);
	CREATE TABLE IF NOT EXISTS Relationship (
		source_id TEXT,
		target_id TEXT,
		type TEXT,
		details TEXT,
		PRIMARY KEY (source_id, target_id)
	);
	CREATE TABLE RelationshipByType (
		source_id TEXT NOT NULL REFERENCES Person(id),
		target_id TEXT NOT NULL REFERENCES Person(id),
		type TEXT NOT NULL,
		details TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (source_id, target_id, type)
	);
	INSERT INTO RelationshipByType (source_id, target_id, type, details)
		SELECT source_id, target_id, COALESCE(type, ''), COALESCE(details, '') FROM Relationship;
	DROP TABLE Relationship;
	ALTER TABLE RelationshipByType RENAME TO Relationship;
	`,
}

func (s *Store) migrate(ctx context.Context) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting migration: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	version, err := schemaVersion(ctx, tx)
	if err != nil {
		return err
	}
	for ; version < len(migrations); version++ {
		if err := applyMigration(ctx, tx, version); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func schemaVersion(ctx context.Context, tx *sql.Tx) (int, error) {
	var version int
	if err := tx.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil {
		return 0, fmt.Errorf("reading schema version: %w", err)
	}
	return version, nil
}

func applyMigration(ctx context.Context, tx *sql.Tx, fromVersion int) error {
	if _, err := tx.ExecContext(ctx, migrations[fromVersion]); err != nil {
		return fmt.Errorf("migrating schema to v%d: %w", fromVersion+1, err)
	}
	// PRAGMA does not accept bound parameters.
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version = %d`, fromVersion+1)); err != nil {
		return fmt.Errorf("saving schema version %d: %w", fromVersion+1, err)
	}
	return nil
}
