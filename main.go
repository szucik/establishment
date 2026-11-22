package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"establishment/v1/internal/auth"
	"establishment/v1/internal/httpapi"
	"establishment/v1/internal/neo4jstore"
	"establishment/v1/internal/password"
	"establishment/v1/internal/person"
	"establishment/v1/internal/sqlitestore"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type repository interface {
	person.Repository
	auth.UserRepository
	auth.SessionRepository
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	store, closeStore, err := openRepository(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeStore()) }()

	server := newServer(cfg, store)
	return serveUntilStopped(ctx, server)
}

func openRepository(ctx context.Context, cfg config) (repository, func() error, error) {
	if cfg.storage == storageNeo4j {
		store, err := neo4jstore.Open(ctx, cfg.neo4j)
		if err != nil {
			return nil, nil, err
		}
		return store, func() error { return store.Close(context.Background()) }, nil
	}
	store, err := sqlitestore.Open(ctx, cfg.sqlitePath)
	if err != nil {
		return nil, nil, err
	}
	return store, store.Close, nil
}

func newServer(cfg config, store repository) *http.Server {
	persons := person.NewService(store)
	accounts := auth.NewService(store, store, password.BcryptHasher{})
	handler := httpapi.NewHandler(persons, accounts, httpapi.Config{
		AllowedOrigins: cfg.allowedOrigins,
		SecureCookies:  cfg.secureCookies,
		StaticDir:      cfg.staticDir,
	})
	return &http.Server{
		Addr:              cfg.address,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func serveUntilStopped(ctx context.Context, server *http.Server) error {
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
