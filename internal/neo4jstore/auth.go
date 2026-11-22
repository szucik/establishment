package neo4jstore

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"establishment/v1/internal/auth"
)

const returnUser = ` RETURN u.id AS id, u.login AS login, COALESCE(u.email, '') AS email, u.password AS password`

func (s *Store) AddUser(ctx context.Context, u auth.User) error {
	_, err := s.write(ctx,
		`CREATE (:User {id: $id, login: $login, email: $email, password: $password})`,
		map[string]any{"id": u.ID, "login": u.Login, "email": u.Email, "password": u.PasswordHash})
	if isConstraintViolation(err) {
		return auth.ErrUserAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("creating user %q: %w", u.Login, err)
	}
	return nil
}

func (s *Store) FindUserByLogin(ctx context.Context, login string) (auth.User, error) {
	return s.findUser(ctx, `MATCH (u:User {login: $value})`, login)
}

func (s *Store) FindUserByID(ctx context.Context, id string) (auth.User, error) {
	return s.findUser(ctx, `MATCH (u:User {id: $value})`, id)
}

func (s *Store) findUser(ctx context.Context, match string, value string) (auth.User, error) {
	records, err := s.read(ctx, match+returnUser, map[string]any{"value": value})
	if err != nil {
		return auth.User{}, fmt.Errorf("querying user: %w", err)
	}
	if len(records) == 0 {
		return auth.User{}, auth.ErrUserNotFound
	}
	return userFrom(records[0])
}

func (s *Store) AddSession(ctx context.Context, session auth.Session) error {
	_, err := s.write(ctx,
		`CREATE (:Session {id: $id, userId: $userId, expiresAt: $expiresAt})`,
		map[string]any{"id": session.ID, "userId": session.UserID, "expiresAt": session.ExpiresAt.Unix()})
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	return nil
}

func (s *Store) FindSession(ctx context.Context, id string) (auth.Session, error) {
	records, err := s.read(ctx,
		`MATCH (s:Session {id: $id}) RETURN s.id AS id, s.userId AS userId, s.expiresAt AS expiresAt`,
		map[string]any{"id": id})
	if err != nil {
		return auth.Session{}, fmt.Errorf("querying session: %w", err)
	}
	if len(records) == 0 {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return sessionFrom(records[0])
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	if _, err := s.write(ctx, `MATCH (s:Session {id: $id}) DELETE s`, map[string]any{"id": id}); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

func userFrom(record *neo4j.Record) (auth.User, error) {
	reader := &recordReader{record: record}
	u := auth.User{
		ID:           reader.string("id"),
		Login:        reader.string("login"),
		Email:        reader.string("email"),
		PasswordHash: reader.string("password"),
	}
	return u, reader.err
}

func sessionFrom(record *neo4j.Record) (auth.Session, error) {
	reader := &recordReader{record: record}
	s := auth.Session{
		ID:        reader.string("id"),
		UserID:    reader.string("userId"),
		ExpiresAt: time.Unix(reader.int64("expiresAt"), 0),
	}
	return s, reader.err
}
