package models

import "context"

type Database interface {
	AddPerson(ctx context.Context, person Person) error
	GetPerson(ctx context.Context, id string) (Person, error)
	GetPersons(ctx context.Context) ([]Person, error)
	AddRelationship(ctx context.Context, rel Relationship) error
	GetGraph(ctx context.Context) (Graph, error)
	AddUser(ctx context.Context, user User) error
	GetUserByLogin(ctx context.Context, login string) (User, error)
	GetUserByID(ctx context.Context, id string) (User, error)
	CreateSession(ctx context.Context, session Session) error
	GetSession(ctx context.Context, sessionID string) (Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type Person struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Occupation  string `json:"occupation"`
	ImageURL    string `json:"image_url"`
	Twitter     string `json:"twitter"`
	Description string `json:"description"`
}

type Relationship struct {
	From    string `json:"source_id"`
	To      string `json:"target_id"`
	Type    string `json:"type"`
	Details string `json:"details"`
}

type Graph struct {
	Nodes []Person       `json:"nodes"`
	Edges []Relationship `json:"edges"`
}

type User struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Session struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	ExpiresAt int64  `json:"expiresAt"`
}
