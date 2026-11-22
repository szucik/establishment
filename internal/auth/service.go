package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

const (
	SessionLifetime = 24 * time.Hour
	randomIDBytes   = 32
)

type UserRepository interface {
	AddUser(ctx context.Context, u User) error
	FindUserByLogin(ctx context.Context, login string) (User, error)
	FindUserByID(ctx context.Context, id string) (User, error)
}

type SessionRepository interface {
	AddSession(ctx context.Context, s Session) error
	FindSession(ctx context.Context, id string) (Session, error)
	DeleteSession(ctx context.Context, id string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Matches(hash, password string) bool
}

type Service struct {
	users    UserRepository
	sessions SessionRepository
	hasher   PasswordHasher
	now      func() time.Time
}

type Option func(*Service)

func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

func NewService(users UserRepository, sessions SessionRepository, hasher PasswordHasher, options ...Option) *Service {
	service := &Service{users: users, sessions: sessions, hasher: hasher, now: time.Now}
	for _, apply := range options {
		apply(service)
	}
	return service
}

func (s *Service) Register(ctx context.Context, registration Registration) error {
	if err := registration.Validate(); err != nil {
		return err
	}
	passwordHash, err := s.hasher.Hash(registration.Password)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}
	userID, err := newRandomID()
	if err != nil {
		return err
	}
	return s.users.AddUser(ctx, User{
		ID:           userID,
		Login:        registration.Login,
		Email:        registration.Email,
		PasswordHash: passwordHash,
	})
}

func (s *Service) Login(ctx context.Context, login, password string) (Session, error) {
	user, err := s.userWithMatchingPassword(ctx, login, password)
	if err != nil {
		return Session{}, err
	}
	session, err := s.newSessionFor(user)
	if err != nil {
		return Session{}, err
	}
	if err := s.sessions.AddSession(ctx, session); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.sessions.DeleteSession(ctx, sessionID)
}

func (s *Service) Authenticate(ctx context.Context, sessionID string) (User, error) {
	session, err := s.sessions.FindSession(ctx, sessionID)
	if err != nil {
		return User{}, err
	}
	if session.IsExpiredAt(s.now()) {
		return User{}, ErrSessionExpired
	}
	return s.users.FindUserByID(ctx, session.UserID)
}

func (s *Service) userWithMatchingPassword(ctx context.Context, login, password string) (User, error) {
	user, err := s.users.FindUserByLogin(ctx, login)
	if errors.Is(err, ErrUserNotFound) {
		return User{}, ErrInvalidCredentials
	}
	if err != nil {
		return User{}, err
	}
	if !s.hasher.Matches(user.PasswordHash, password) {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) newSessionFor(user User) (Session, error) {
	sessionID, err := newRandomID()
	if err != nil {
		return Session{}, err
	}
	return Session{ID: sessionID, UserID: user.ID, ExpiresAt: s.now().Add(SessionLifetime)}, nil
}

func newRandomID() (string, error) {
	bytes := make([]byte, randomIDBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating random id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
