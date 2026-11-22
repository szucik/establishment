package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"establishment/v1/internal/auth"
)

const selectUsers = `SELECT id, login, email, password FROM User`

func (s *Store) AddUser(ctx context.Context, u auth.User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO User (id, login, email, password) VALUES (?, ?, ?, ?)`,
		u.ID, u.Login, u.Email, u.PasswordHash)
	if isDuplicate(err) {
		return auth.ErrUserAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("inserting user %q: %w", u.Login, err)
	}
	return nil
}

func (s *Store) FindUserByLogin(ctx context.Context, login string) (auth.User, error) {
	return s.findUser(ctx, `login = ?`, login)
}

func (s *Store) FindUserByID(ctx context.Context, id string) (auth.User, error) {
	return s.findUser(ctx, `id = ?`, id)
}

func (s *Store) findUser(ctx context.Context, condition string, value string) (auth.User, error) {
	var u auth.User
	err := s.db.QueryRowContext(ctx, selectUsers+` WHERE `+condition, value).
		Scan(&u.ID, &u.Login, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrUserNotFound
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("querying user: %w", err)
	}
	return u, nil
}

func (s *Store) AddSession(ctx context.Context, session auth.Session) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO Session (id, user_id, expires_at) VALUES (?, ?, ?)`,
		session.ID, session.UserID, session.ExpiresAt.Unix())
	if err != nil {
		return fmt.Errorf("inserting session: %w", err)
	}
	return nil
}

func (s *Store) FindSession(ctx context.Context, id string) (auth.Session, error) {
	var session auth.Session
	var expiresAtUnix int64
	err := s.db.QueryRowContext(ctx, `SELECT id, user_id, expires_at FROM Session WHERE id = ?`, id).
		Scan(&session.ID, &session.UserID, &expiresAtUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	if err != nil {
		return auth.Session{}, fmt.Errorf("querying session: %w", err)
	}
	session.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return session, nil
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM Session WHERE id = ?`, id); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}
