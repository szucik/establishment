package auth

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrMissingCredentials = errors.New("login, email and password are required")
	ErrUserAlreadyExists  = errors.New("user with this login or email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

type User struct {
	ID           string
	Login        string
	Email        string
	PasswordHash string
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}

func (s Session) IsExpiredAt(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}

type Registration struct {
	Login    string
	Email    string
	Password string
}

func (r Registration) Validate() error {
	if isBlank(r.Login) || isBlank(r.Email) || r.Password == "" {
		return ErrMissingCredentials
	}
	return nil
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}
