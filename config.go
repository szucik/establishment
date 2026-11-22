package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"establishment/v1/internal/neo4jstore"
)

const (
	storageSQLite = "sqlite"
	storageNeo4j  = "neo4j"
)

type config struct {
	address        string
	storage        string
	sqlitePath     string
	neo4j          neo4jstore.Credentials
	allowedOrigins []string
	secureCookies  bool
	staticDir      string
}

func loadConfig() (config, error) {
	cfg := config{
		address:        envOrDefault("ADDR", ":8080"),
		storage:        envOrDefault("STORAGE", storageSQLite),
		sqlitePath:     envOrDefault("SQLITE_PATH", "./data.db"),
		allowedOrigins: strings.Split(envOrDefault("ALLOWED_ORIGINS", "http://localhost:5500,http://127.0.0.1:5500"), ","),
		secureCookies:  os.Getenv("COOKIE_SECURE") == "true",
		staticDir:      envOrDefault("STATIC_DIR", "static"),
		neo4j: neo4jstore.Credentials{
			URI:      envOrDefault("NEO4J_URI", "neo4j://localhost:7687"),
			User:     envOrDefault("NEO4J_USER", "neo4j"),
			Password: os.Getenv("NEO4J_PASSWORD"),
		},
	}
	return cfg, cfg.validate()
}

func (c config) validate() error {
	switch c.storage {
	case storageSQLite:
		return nil
	case storageNeo4j:
		if c.neo4j.Password == "" {
			return errors.New("NEO4J_PASSWORD is required when STORAGE=neo4j")
		}
		return nil
	default:
		return fmt.Errorf("unknown STORAGE %q (use %q or %q)", c.storage, storageSQLite, storageNeo4j)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
