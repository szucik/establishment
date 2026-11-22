package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"establishment/v1/internal/auth"
	"establishment/v1/internal/httpapi"
	"establishment/v1/internal/person"
)

const validSessionID = "valid-session"

type stubPersons struct {
	addPersonErr error
	graph        person.Graph
	graphErr     error
}

func (s *stubPersons) AddPerson(context.Context, person.Person) error { return s.addPersonErr }
func (s *stubPersons) Person(context.Context, string) (person.Person, error) {
	return person.Person{}, person.ErrNotFound
}
func (s *stubPersons) Persons(context.Context) ([]person.Person, error)           { return nil, nil }
func (s *stubPersons) AddRelationship(context.Context, person.Relationship) error { return nil }
func (s *stubPersons) Graph(context.Context) (person.Graph, error)                { return s.graph, s.graphErr }

type stubAccounts struct{}

func (stubAccounts) Register(context.Context, auth.Registration) error { return nil }
func (stubAccounts) Login(context.Context, string, string) (auth.Session, error) {
	return auth.Session{}, auth.ErrInvalidCredentials
}
func (stubAccounts) Logout(context.Context, string) error { return nil }
func (stubAccounts) Authenticate(_ context.Context, sessionID string) (auth.User, error) {
	if sessionID != validSessionID {
		return auth.User{}, auth.ErrSessionNotFound
	}
	return auth.User{ID: "u1", Login: "jan"}, nil
}

func serve(persons *stubPersons, request *http.Request) *httptest.ResponseRecorder {
	handler := httpapi.NewHandler(persons, stubAccounts{}, httpapi.Config{AllowedOrigins: []string{"http://localhost:5500"}})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func postPersonRequest(sessionID string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/person", strings.NewReader(`{"id":"1","name":"Jan"}`))
	if sessionID != "" {
		request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	}
	return request
}

func TestPostPerson_ValidSession_ReturnsCreated(t *testing.T) {
	response := serve(&stubPersons{}, postPersonRequest(validSessionID))

	if response.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", response.Code, http.StatusCreated)
	}
}

func TestPostPerson_NoSessionCookie_ReturnsUnauthorized(t *testing.T) {
	response := serve(&stubPersons{}, postPersonRequest(""))

	if response.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestPostPerson_DuplicateID_ReturnsConflict(t *testing.T) {
	response := serve(&stubPersons{addPersonErr: person.ErrAlreadyExists}, postPersonRequest(validSessionID))

	if response.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

func TestGetPerson_UnknownID_ReturnsNotFound(t *testing.T) {
	response := serve(&stubPersons{}, httptest.NewRequest(http.MethodGet, "/person/ghost", nil))

	if response.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestGetGraph_EmptyGraph_ReturnsEmptyArraysNotNull(t *testing.T) {
	response := serve(&stubPersons{}, httptest.NewRequest(http.MethodGet, "/graph", nil))

	if body := strings.TrimSpace(response.Body.String()); body != `{"nodes":[],"edges":[]}` {
		t.Errorf("body = %s, want empty arrays", body)
	}
}

func TestGetGraph_StorageFailure_HidesInternalErrorDetails(t *testing.T) {
	persons := &stubPersons{graphErr: errors.New("disk on fire at /var/lib/data.db")}

	response := serve(persons, httptest.NewRequest(http.MethodGet, "/graph", nil))

	if response.Code != http.StatusInternalServerError || strings.Contains(response.Body.String(), "disk") {
		t.Errorf("got %d %q, want 500 without internal details", response.Code, response.Body.String())
	}
}

func TestCheckSession_ValidSession_ReturnsLogin(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/check-session", nil)
	request.AddCookie(&http.Cookie{Name: "session_id", Value: validSessionID})

	response := serve(&stubPersons{}, request)

	if body := strings.TrimSpace(response.Body.String()); body != `{"login":"jan"}` {
		t.Errorf("body = %s, want login jan", body)
	}
}

func TestPreflight_AllowedOrigin_ReturnsCORSHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodOptions, "/person", nil)
	request.Header.Set("Origin", "http://localhost:5500")

	response := serve(&stubPersons{}, request)

	if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5500" {
		t.Errorf("got %d with origin %q, want 204 allowing localhost:5500", response.Code, response.Header().Get("Access-Control-Allow-Origin"))
	}
}
