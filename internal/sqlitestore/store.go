package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

// SQLite allows a single writer; one connection avoids "database is locked" errors.
const maxOpenConnections = 1

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite3", dataSourceName(path))
	if err != nil {
		return nil, fmt.Errorf("opening sqlite %q: %w", path, err)
	}
	db.SetMaxOpenConns(maxOpenConnections)
	store := &Store{db: db}
	if err := store.prepare(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func dataSourceName(path string) string {
	return fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", path)
}

func (s *Store) prepare(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("connecting to sqlite: %w", err)
	}
	return s.migrate(ctx)
}

func isConstraintViolation(err error, code sqlite3.ErrNoExtended) bool {
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == code
}

func isDuplicate(err error) bool {
	return isConstraintViolation(err, sqlite3.ErrConstraintPrimaryKey) ||
		isConstraintViolation(err, sqlite3.ErrConstraintUnique)
}

func closeRows(rows *sql.Rows, err *error) {
	*err = errors.Join(*err, rows.Close())
}
