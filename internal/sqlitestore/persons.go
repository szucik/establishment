package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"

	"establishment/v1/internal/person"
)

const selectPersons = `
	SELECT id, name, COALESCE(occupation, ''), COALESCE(image_url, ''), COALESCE(twitter, ''), COALESCE(description, '')
	FROM Person`

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *Store) AddPerson(ctx context.Context, p person.Person) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO Person (id, name, occupation, image_url, twitter, description) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Occupation, p.ImageURL, p.Twitter, p.Description)
	if isDuplicate(err) {
		return person.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("inserting person %q: %w", p.ID, err)
	}
	return nil
}

func (s *Store) FindPersonByID(ctx context.Context, id string) (person.Person, error) {
	found, err := scanPerson(s.db.QueryRowContext(ctx, selectPersons+` WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return person.Person{}, person.ErrNotFound
	}
	if err != nil {
		return person.Person{}, fmt.Errorf("querying person %q: %w", id, err)
	}
	return found, nil
}

func (s *Store) FindAllPersons(ctx context.Context) (persons []person.Person, err error) {
	rows, err := s.db.QueryContext(ctx, selectPersons+` ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("querying persons: %w", err)
	}
	defer closeRows(rows, &err)

	persons = []person.Person{}
	for rows.Next() {
		p, scanErr := scanPerson(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning person: %w", scanErr)
		}
		persons = append(persons, p)
	}
	return persons, rows.Err()
}

func (s *Store) SaveRelationship(ctx context.Context, r person.Relationship) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO Relationship (source_id, target_id, type, details) VALUES (?, ?, ?, ?)
		 ON CONFLICT (source_id, target_id, type) DO UPDATE SET details = excluded.details`,
		r.SourceID, r.TargetID, r.Type, r.Details)
	if isConstraintViolation(err, sqlite3.ErrConstraintForeignKey) {
		return person.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("saving relationship %s->%s: %w", r.SourceID, r.TargetID, err)
	}
	return nil
}

func (s *Store) FindAllRelationships(ctx context.Context) (relationships []person.Relationship, err error) {
	rows, err := s.db.QueryContext(ctx, `SELECT source_id, target_id, type, details FROM Relationship`)
	if err != nil {
		return nil, fmt.Errorf("querying relationships: %w", err)
	}
	defer closeRows(rows, &err)

	relationships = []person.Relationship{}
	for rows.Next() {
		var r person.Relationship
		if scanErr := rows.Scan(&r.SourceID, &r.TargetID, &r.Type, &r.Details); scanErr != nil {
			return nil, fmt.Errorf("scanning relationship: %w", scanErr)
		}
		relationships = append(relationships, r)
	}
	return relationships, rows.Err()
}

func scanPerson(row rowScanner) (person.Person, error) {
	var p person.Person
	err := row.Scan(&p.ID, &p.Name, &p.Occupation, &p.ImageURL, &p.Twitter, &p.Description)
	return p, err
}
