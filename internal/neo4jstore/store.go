package neo4jstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const constraintViolationCode = "Neo.ClientError.Schema.ConstraintValidationFailed"

var uniquenessConstraints = []string{
	`CREATE CONSTRAINT person_id_unique IF NOT EXISTS FOR (p:Person) REQUIRE p.id IS UNIQUE`,
	`CREATE CONSTRAINT user_id_unique IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE`,
	`CREATE CONSTRAINT user_login_unique IF NOT EXISTS FOR (u:User) REQUIRE u.login IS UNIQUE`,
	`CREATE CONSTRAINT user_email_unique IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE`,
	`CREATE CONSTRAINT session_id_unique IF NOT EXISTS FOR (s:Session) REQUIRE s.id IS UNIQUE`,
}

type Credentials struct {
	URI      string
	User     string
	Password string
}

type Store struct {
	driver neo4j.DriverWithContext
}

func Open(ctx context.Context, credentials Credentials) (*Store, error) {
	driver, err := neo4j.NewDriverWithContext(credentials.URI, neo4j.BasicAuth(credentials.User, credentials.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("creating neo4j driver for %s: %w", credentials.URI, err)
	}
	store := &Store{driver: driver}
	if err := store.prepare(ctx); err != nil {
		return nil, errors.Join(err, driver.Close(ctx))
	}
	return store, nil
}

func (s *Store) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

func (s *Store) prepare(ctx context.Context) error {
	if err := s.driver.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("connecting to neo4j: %w", err)
	}
	for _, constraint := range uniquenessConstraints {
		if _, err := s.write(ctx, constraint, nil); err != nil {
			return fmt.Errorf("creating constraint (check for duplicate data): %w", err)
		}
	}
	return nil
}

func (s *Store) write(ctx context.Context, query string, parameters map[string]any) ([]*neo4j.Record, error) {
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, parameters, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}
	return result.Records, nil
}

func (s *Store) read(ctx context.Context, query string, parameters map[string]any) ([]*neo4j.Record, error) {
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, parameters, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithReadersRouting())
	if err != nil {
		return nil, err
	}
	return result.Records, nil
}

func isConstraintViolation(err error) bool {
	var neo4jErr *neo4j.Neo4jError
	return errors.As(err, &neo4jErr) && neo4jErr.Code == constraintViolationCode
}

// recordReader keeps the first decoding error so mapping code stays linear.
type recordReader struct {
	record *neo4j.Record
	err    error
}

func (r *recordReader) string(key string) string {
	return readValue[string](r, key)
}

func (r *recordReader) int64(key string) int64 {
	return readValue[int64](r, key)
}

func readValue[T neo4j.RecordValue](r *recordReader, key string) T {
	var zero T
	if r.err != nil {
		return zero
	}
	value, _, err := neo4j.GetRecordValue[T](r.record, key)
	if err != nil {
		r.err = fmt.Errorf("reading %q: %w", key, err)
	}
	return value
}
