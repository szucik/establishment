package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"establishment/v1/internal/auth"
)

type fakeUsers struct{ byID map[string]auth.User }

func (f *fakeUsers) AddUser(_ context.Context, u auth.User) error {
	for _, existing := range f.byID {
		if existing.Login == u.Login || existing.Email == u.Email {
			return auth.ErrUserAlreadyExists
		}
	}
	f.byID[u.ID] = u
	return nil
}

func (f *fakeUsers) FindUserByLogin(_ context.Context, login string) (auth.User, error) {
	for _, u := range f.byID {
		if u.Login == login {
			return u, nil
		}
	}
	return auth.User{}, auth.ErrUserNotFound
}

func (f *fakeUsers) FindUserByID(_ context.Context, id string) (auth.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return auth.User{}, auth.ErrUserNotFound
	}
	return u, nil
}

type fakeSessions struct{ byID map[string]auth.Session }

func (f *fakeSessions) AddSession(_ context.Context, s auth.Session) error {
	f.byID[s.ID] = s
	return nil
}

func (f *fakeSessions) FindSession(_ context.Context, id string) (auth.Session, error) {
	s, ok := f.byID[id]
	if !ok {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return s, nil
}

func (f *fakeSessions) DeleteSession(_ context.Context, id string) error {
	delete(f.byID, id)
	return nil
}

type plainTextHasher struct{}

func (plainTextHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (plainTextHasher) Matches(hash, password string) bool   { return hash == "hashed:"+password }

type stoppedClock struct{ now time.Time }

func (c *stoppedClock) Now() time.Time { return c.now }

var (
	startTime    = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	registration = auth.Registration{Login: "jan", Email: "jan@example.com", Password: "secret"}
)

func newServiceWithRegisteredUser(t *testing.T, clock *stoppedClock) *auth.Service {
	t.Helper()
	service := auth.NewService(
		&fakeUsers{byID: map[string]auth.User{}},
		&fakeSessions{byID: map[string]auth.Session{}},
		plainTextHasher{},
		auth.WithClock(clock.Now),
	)
	if err := service.Register(context.Background(), registration); err != nil {
		t.Fatalf("Register() = %v", err)
	}
	return service
}

func TestRegister_MissingPassword_ReturnsMissingCredentials(t *testing.T) {
	service := newServiceWithRegisteredUser(t, &stoppedClock{now: startTime})

	err := service.Register(context.Background(), auth.Registration{Login: "anna", Email: "anna@example.com"})

	if !errors.Is(err, auth.ErrMissingCredentials) {
		t.Errorf("Register() = %v, want %v", err, auth.ErrMissingCredentials)
	}
}

func TestRegister_DuplicateLogin_ReturnsUserAlreadyExists(t *testing.T) {
	service := newServiceWithRegisteredUser(t, &stoppedClock{now: startTime})

	err := service.Register(context.Background(), registration)

	if !errors.Is(err, auth.ErrUserAlreadyExists) {
		t.Errorf("Register() = %v, want %v", err, auth.ErrUserAlreadyExists)
	}
}

func TestLogin_WrongPassword_ReturnsInvalidCredentials(t *testing.T) {
	service := newServiceWithRegisteredUser(t, &stoppedClock{now: startTime})

	_, err := service.Login(context.Background(), registration.Login, "wrong")

	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("Login() = %v, want %v", err, auth.ErrInvalidCredentials)
	}
}

func TestLogin_UnknownLogin_ReturnsInvalidCredentials(t *testing.T) {
	service := newServiceWithRegisteredUser(t, &stoppedClock{now: startTime})

	_, err := service.Login(context.Background(), "ghost", registration.Password)

	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("Login() = %v, want %v", err, auth.ErrInvalidCredentials)
	}
}

func TestLogin_ValidCredentials_SessionExpiresAfterLifetime(t *testing.T) {
	service := newServiceWithRegisteredUser(t, &stoppedClock{now: startTime})

	session, err := service.Login(context.Background(), registration.Login, registration.Password)

	if err != nil || !session.ExpiresAt.Equal(startTime.Add(auth.SessionLifetime)) {
		t.Errorf("Login() = %+v, %v; want session expiring at %v", session, err, startTime.Add(auth.SessionLifetime))
	}
}

func TestAuthenticate_ActiveSession_ReturnsLoggedInUser(t *testing.T) {
	service := newServiceWithRegisteredUser(t, &stoppedClock{now: startTime})
	session, _ := service.Login(context.Background(), registration.Login, registration.Password)

	user, err := service.Authenticate(context.Background(), session.ID)

	if err != nil || user.Login != registration.Login {
		t.Errorf("Authenticate() = %+v, %v; want user %q", user, err, registration.Login)
	}
}

func TestAuthenticate_ExpiredSession_ReturnsSessionExpired(t *testing.T) {
	clock := &stoppedClock{now: startTime}
	service := newServiceWithRegisteredUser(t, clock)
	session, _ := service.Login(context.Background(), registration.Login, registration.Password)
	clock.now = startTime.Add(auth.SessionLifetime)

	_, err := service.Authenticate(context.Background(), session.ID)

	if !errors.Is(err, auth.ErrSessionExpired) {
		t.Errorf("Authenticate() = %v, want %v", err, auth.ErrSessionExpired)
	}
}

func TestAuthenticate_AfterLogout_ReturnsSessionNotFound(t *testing.T) {
	service := newServiceWithRegisteredUser(t, &stoppedClock{now: startTime})
	session, _ := service.Login(context.Background(), registration.Login, registration.Password)
	if err := service.Logout(context.Background(), session.ID); err != nil {
		t.Fatalf("Logout() = %v", err)
	}

	_, err := service.Authenticate(context.Background(), session.ID)

	if !errors.Is(err, auth.ErrSessionNotFound) {
		t.Errorf("Authenticate() = %v, want %v", err, auth.ErrSessionNotFound)
	}
}
