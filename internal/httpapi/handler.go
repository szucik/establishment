package httpapi

import (
	"context"
	"net/http"

	"establishment/v1/internal/auth"
	"establishment/v1/internal/person"
)

type personCatalog interface {
	AddPerson(ctx context.Context, p person.Person) error
	Person(ctx context.Context, id string) (person.Person, error)
	Persons(ctx context.Context) ([]person.Person, error)
	AddRelationship(ctx context.Context, r person.Relationship) error
	Graph(ctx context.Context) (person.Graph, error)
}

type sessionAuthenticator interface {
	Authenticate(ctx context.Context, sessionID string) (auth.User, error)
}

type accountManager interface {
	sessionAuthenticator
	Register(ctx context.Context, registration auth.Registration) error
	Login(ctx context.Context, login, password string) (auth.Session, error)
	Logout(ctx context.Context, sessionID string) error
}

type Config struct {
	AllowedOrigins []string
	SecureCookies  bool
	StaticDir      string
}

func NewHandler(persons personCatalog, accounts accountManager, config Config) http.Handler {
	personRoutes := personHandlers{persons: persons}
	accountRoutes := accountHandlers{accounts: accounts, secureCookies: config.SecureCookies}
	requireLogin := requireSession(accounts)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /person/{id}", personRoutes.getPerson)
	mux.HandleFunc("GET /persons", personRoutes.listPersons)
	mux.HandleFunc("GET /graph", personRoutes.getGraph)
	mux.Handle("POST /person", requireLogin(http.HandlerFunc(personRoutes.postPerson)))
	mux.Handle("POST /relationship", requireLogin(http.HandlerFunc(personRoutes.postRelationship)))

	mux.HandleFunc("POST /register", accountRoutes.register)
	mux.HandleFunc("POST /login", accountRoutes.login)
	mux.HandleFunc("POST /logout", accountRoutes.logout)
	mux.Handle("GET /check-session", requireLogin(http.HandlerFunc(accountRoutes.checkSession)))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticDir))))
	mux.Handle("GET /{$}", http.RedirectHandler("/static/", http.StatusFound))

	return allowOrigins(config.AllowedOrigins)(mux)
}
