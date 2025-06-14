package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"establishment/v1/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrNoSuchPerson        = errors.New("no such person")
	ErrUserExists          = errors.New("user already exists")
	ErrNoSuchSession       = errors.New("no such session")
	ErrInvalidRelationship = errors.New("source and target IDs must be different")
)

// SQLiteDB wraps a SQLite database connection.
type SQLiteDB struct {
	*sql.DB
}

// ConnectToSQLite opens a connection to the SQLite database in data.db.
func ConnectToSQLite(ctx context.Context) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Printf("Failed to connect to SQLite: %v", err)
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		log.Printf("Failed to verify SQLite connectivity: %v", err)
		return nil, fmt.Errorf("failed to verify SQLite connectivity: %w", err)
	}

	// Limit connections for SQLite
	db.SetMaxOpenConns(1)

	log.Println("Successfully connected to SQLite")
	return &SQLiteDB{DB: db}, nil
}

// InitSchema creates the database schema if it does not exist.
func InitSchema(ctx context.Context, db *SQLiteDB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS Person (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			occupation TEXT,
			image_url TEXT,
			twitter TEXT,
			description TEXT
		);
		CREATE TABLE IF NOT EXISTS Relationship (
			source_id TEXT,
			target_id TEXT,
			type TEXT,
			details TEXT,
			PRIMARY KEY (source_id, target_id),
			FOREIGN KEY (source_id) REFERENCES Person(id),
			FOREIGN KEY (target_id) REFERENCES Person(id)
		);
		CREATE TABLE IF NOT EXISTS User (
			id TEXT PRIMARY KEY,
			login TEXT UNIQUE,
			email TEXT UNIQUE,
			password TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS Session (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			expires_at INTEGER,
			FOREIGN KEY (user_id) REFERENCES User(id)
		);
	`)
	if err != nil {
		log.Printf("Failed to initialize SQLite schema: %v", err)
		return fmt.Errorf("failed to initialize SQLite schema: %w", err)
	}
	log.Println("SQLite schema initialized successfully")
	return nil
}

// AddPerson adds a person to the database.
func (db *SQLiteDB) AddPerson(ctx context.Context, person models.Person) error {
	log.Printf("Adding person: id=%s, name=%s, occupation=%s", person.ID, person.Name, person.Occupation)

	_, err := db.ExecContext(ctx,
		`INSERT INTO Person (id, name, occupation, image_url, twitter, description)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		person.ID, person.Name, person.Occupation, person.ImageURL, person.Twitter, person.Description)
	if err != nil {
		log.Printf("Failed to add person: %v", err)
		return fmt.Errorf("failed to add person: %w", err)
	}
	log.Println("Person added successfully")
	return nil
}

// GetPerson retrieves a person by their ID.
func (db *SQLiteDB) GetPerson(ctx context.Context, id string) (models.Person, error) {
	var person models.Person
	err := db.QueryRowContext(ctx,
		`SELECT id, name, occupation, image_url, twitter, description
		 FROM Person WHERE id = ?`, id).
		Scan(&person.ID, &person.Name, &person.Occupation, &person.ImageURL, &person.Twitter, &person.Description)
	if err == sql.ErrNoRows {
		log.Printf("Person not found for ID: %s", id)
		return models.Person{}, ErrNoSuchPerson
	}
	if err != nil {
		log.Printf("Failed to query person for ID %s: %v", id, err)
		return models.Person{}, fmt.Errorf("failed to query person: %w", err)
	}
	return person, nil
}

// GetPersons retrieves all persons from the database.
func (db *SQLiteDB) GetPersons(ctx context.Context) ([]models.Person, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, occupation, image_url, twitter, description
		 FROM Person`)
	if err != nil {
		log.Printf("Failed to query persons: %v", err)
		return nil, fmt.Errorf("failed to query persons: %w", err)
	}
	defer rows.Close()

	var persons []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Occupation, &p.ImageURL, &p.Twitter, &p.Description); err != nil {
			log.Printf("Failed to scan person: %v", err)
			return nil, fmt.Errorf("failed to scan person: %w", err)
		}
		persons = append(persons, p)
	}
	log.Printf("Returning %d persons", len(persons))
	return persons, nil
}

// AddRelationship adds a relationship between two persons.
func (db *SQLiteDB) AddRelationship(ctx context.Context, rel models.Relationship) error {
	if rel.From == rel.To {
		log.Printf("Invalid relationship: source_id=%s and target_id=%s are the same", rel.From, rel.To)
		return ErrInvalidRelationship
	}

	log.Printf("Verifying persons for relationship: source_id=%s, target_id=%s", rel.From, rel.To)

	// Verify existence of persons
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM Person WHERE id IN (?, ?)`, rel.From, rel.To).
		Scan(&count)
	if err != nil {
		log.Printf("Failed to verify persons: %v", err)
		return fmt.Errorf("failed to verify persons for relationship: %w", err)
	}
	if count != 2 {
		log.Printf("One or both persons not found: source_id=%s, target_id=%s", rel.From, rel.To)
		return fmt.Errorf("one or both persons not found: source_id=%s, target_id=%s", rel.From, rel.To)
	}

	log.Printf("Adding relationship: source_id=%s, target_id=%s, type=%s, details=%s", rel.From, rel.To, rel.Type, rel.Details)

	_, err = db.ExecContext(ctx,
		`INSERT INTO Relationship (source_id, target_id, type, details)
		 VALUES (?, ?, ?, ?)`,
		rel.From, rel.To, rel.Type, rel.Details)
	if err != nil {
		log.Printf("Failed to add relationship: %v", err)
		return fmt.Errorf("failed to add relationship: %w", err)
	}
	log.Printf("Relationship added successfully: %s -> %s (%s)", rel.From, rel.To, rel.Type)
	return nil
}

// GetGraph retrieves the graph of persons and their relationships.
func (db *SQLiteDB) GetGraph(ctx context.Context) (models.Graph, error) {
	// Retrieve persons
	rows, err := db.QueryContext(ctx,
		`SELECT id, name, occupation, image_url, twitter, description
		 FROM Person`)
	if err != nil {
		log.Printf("Failed to query persons for graph: %v", err)
		return models.Graph{}, fmt.Errorf("failed to query graph: %w", err)
	}
	defer rows.Close()

	nodes := make(map[string]models.Person)
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Occupation, &p.ImageURL, &p.Twitter, &p.Description); err != nil {
			log.Printf("Failed to scan person for graph: %v", err)
			return models.Graph{}, fmt.Errorf("failed to scan person: %w", err)
		}
		nodes[p.ID] = p
	}

	// Retrieve relationships
	rows, err = db.QueryContext(ctx,
		`SELECT source_id, target_id, type, details
		 FROM Relationship`)
	if err != nil {
		log.Printf("Failed to query relationships for graph: %v", err)
		return models.Graph{}, fmt.Errorf("failed to query graph: %w", err)
	}
	defer rows.Close()

	var edges []models.Relationship
	for rows.Next() {
		var e models.Relationship
		if err := rows.Scan(&e.From, &e.To, &e.Type, &e.Details); err != nil {
			log.Printf("Failed to scan relationship for graph: %v", err)
			return models.Graph{}, fmt.Errorf("failed to scan relationship: %w", err)
		}
		if e.From == e.To {
			log.Printf("Warning: Skipping self-referential edge: source_id=%s, target_id=%s", e.From, e.To)
			continue
		}
		if _, ok := nodes[e.From]; !ok {
			log.Printf("Warning: Skipping edge with missing source_id: %s", e.From)
			continue
		}
		if _, ok := nodes[e.To]; !ok {
			log.Printf("Warning: Skipping edge with missing target_id: %s", e.To)
			continue
		}
		log.Printf("Adding edge: source_id=%s, target_id=%s, type=%s, details=%s", e.From, e.To, e.Type, e.Details)
		edges = append(edges, e)
	}

	nodeList := make([]models.Person, 0, len(nodes))
	for _, node := range nodes {
		nodeList = append(nodeList, node)
	}

	log.Printf("Returning graph: %d nodes, %d edges", len(nodeList), len(edges))
	return models.Graph{Nodes: nodeList, Edges: edges}, nil
}

// AddUser adds a new user to the database.
func (db *SQLiteDB) AddUser(ctx context.Context, user models.User) error {
	// Check if user already exists
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM User WHERE login = ? OR email = ?`, user.Login, user.Email).
		Scan(&count)
	if err != nil {
		log.Printf("Failed to check existing user: %v", err)
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if count > 0 {
		log.Printf("User already exists: login=%s, email=%s", user.Login, user.Email)
		return ErrUserExists
	}

	_, err = db.ExecContext(ctx,
		`INSERT INTO User (id, login, email, password)
		 VALUES (?, ?, ?, ?)`,
		user.ID, user.Login, user.Email, user.Password)
	if err != nil {
		log.Printf("Failed to add user: %v", err)
		return fmt.Errorf("failed to add user: %w", err)
	}
	log.Printf("User added successfully: login=%s", user.Login)
	return nil
}

// GetUserByLogin retrieves a user by their login.
func (db *SQLiteDB) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	var user models.User
	err := db.QueryRowContext(ctx,
		`SELECT id, login, email, password
		 FROM User WHERE login = ?`, login).
		Scan(&user.ID, &user.Login, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		log.Printf("User not found for login: %s", login)
		return models.User{}, fmt.Errorf("user not found")
	}
	if err != nil {
		log.Printf("Failed to query user for login %s: %v", login, err)
		return models.User{}, fmt.Errorf("failed to query user: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func (db *SQLiteDB) GetUserByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	err := db.QueryRowContext(ctx,
		`SELECT id, login, email, password
		 FROM User WHERE id = ?`, id).
		Scan(&user.ID, &user.Login, &user.Email, &user.Password)
	if err == sql.ErrNoRows {
		log.Printf("User not found for ID: %s", id)
		return models.User{}, fmt.Errorf("user not found")
	}
	if err != nil {
		log.Printf("Failed to query user for ID %s: %v", id, err)
		return models.User{}, fmt.Errorf("failed to query user: %w", err)
	}
	return user, nil
}

// CreateSession creates a new user session.
func (db *SQLiteDB) CreateSession(ctx context.Context, session models.Session) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO Session (id, user_id, expires_at)
		 VALUES (?, ?, ?)`,
		session.ID, session.UserID, session.ExpiresAt)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return fmt.Errorf("failed to create session: %w", err)
	}
	log.Printf("Session created successfully: id=%s", session.ID)
	return nil
}

// GetSession retrieves a session by its ID.
func (db *SQLiteDB) GetSession(ctx context.Context, sessionID string) (models.Session, error) {
	var session models.Session
	err := db.QueryRowContext(ctx,
		`SELECT id, user_id, expires_at
		 FROM Session WHERE id = ?`, sessionID).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt)
	if err == sql.ErrNoRows {
		log.Printf("Session not found for ID: %s", sessionID)
		return models.Session{}, ErrNoSuchSession
	}
	if err != nil {
		log.Printf("Failed to query session for ID %s: %v", sessionID, err)
		return models.Session{}, fmt.Errorf("failed to query session: %w", err)
	}
	return session, nil
}

// DeleteSession deletes a session by its ID.
func (db *SQLiteDB) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM Session WHERE id = ?`, sessionID)
	if err != nil {
		log.Printf("Failed to delete session for ID %s: %v", sessionID, err)
		return fmt.Errorf("failed to delete session: %w", err)
	}
	log.Printf("Session deleted successfully: id=%s", sessionID)
	return nil
}
