package database

import (
	"context"
	"errors"
	"fmt"

	"establishment/v1/establishment/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	ErrNoSuchPerson  = errors.New("no such person")
	ErrUserExists    = errors.New("user already exists")
	ErrNoSuchSession = errors.New("no such session")
)

func ConnectToNeo4j(ctx context.Context) (neo4j.DriverWithContext, error) {
	uri := "neo4j://localhost:7687"
	username := "neo4j"
	password := "secretgraph"
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}
	return driver, nil
}

func AddPerson(ctx context.Context, driver neo4j.DriverWithContext, person models.Person) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		`CREATE (p:Person {
            id: $id, 
            name: $name, 
            occupation: $occupation,
            image_url: $image_url,
            twitter: $twitter,
            description: $description
        })`,
		map[string]interface{}{
			"id":          person.ID,
			"name":        person.Name,
			"occupation":  person.Occupation,
			"image_url":   person.ImageURL,
			"twitter":     person.Twitter,
			"description": person.Description,
		})
	if err != nil {
		return fmt.Errorf("failed to add person: %w", err)
	}
	return nil
}

func GetPerson(ctx context.Context, driver neo4j.DriverWithContext, id string) (models.Person, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:Person {id: $id}) 
         RETURN p.id, p.name, p.occupation, p.image_url, p.twitter, p.description`,
		map[string]interface{}{"id": id})
	if err != nil {
		return models.Person{}, fmt.Errorf("failed to query person: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		id, _ := record.Get("p.id")
		name, _ := record.Get("p.name")
		occupation, _ := record.Get("p.occupation")
		imageURL, _ := record.Get("p.image_url")
		twitter, _ := record.Get("p.twitter")
		description, _ := record.Get("p.description")

		return models.Person{
			ID:          id.(string),
			Name:        name.(string),
			Occupation:  occupation.(string),
			ImageURL:    imageURL.(string),
			Twitter:     twitter.(string),
			Description: description.(string),
		}, nil
	}

	return models.Person{}, ErrNoSuchPerson
}

func AddRelationship(ctx context.Context, driver neo4j.DriverWithContext, rel models.Relationship) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		`MATCH (a:Person {id: $from}), (b:Person {id: $to})
         CREATE (a)-[r:RELATIONSHIP {type: $type, details: $details}]->(b)`,
		map[string]interface{}{
			"from":    rel.From,
			"to":      rel.To,
			"type":    rel.Type,
			"details": rel.Details,
		})
	if err != nil {
		return fmt.Errorf("failed to add relationship: %w", err)
	}
	return nil
}

func GetGraph(ctx context.Context, driver neo4j.DriverWithContext) (models.Graph, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:Person)
         OPTIONAL MATCH (p)-[r:RELATIONSHIP]->(q:Person)
         RETURN p.id, p.name, p.occupation, p.image_url, p.twitter, p.description,
                r.type, r.details, q.id as target_id`,
		nil)
	if err != nil {
		return models.Graph{}, fmt.Errorf("failed to query graph: %w", err)
	}

	nodes := make(map[string]models.Person)
	edges := []models.Relationship{}

	for result.Next(ctx) {
		record := result.Record()
		id, _ := record.Get("p.id")
		name, _ := record.Get("p.name")
		occupation, _ := record.Get("p.occupation")
		imageURL, _ := record.Get("p.image_url")
		twitter, _ := record.Get("p.twitter")
		description, _ := record.Get("p.description")

		node := models.Person{
			ID:          id.(string),
			Name:        name.(string),
			Occupation:  occupation.(string),
			ImageURL:    imageURL.(string),
			Twitter:     twitter.(string),
			Description: description.(string),
		}
		nodes[id.(string)] = node

		if targetID, ok := record.Get("target_id"); ok && targetID != nil {
			relType, _ := record.Get("r.type")
			details, _ := record.Get("r.details")
			edges = append(edges, models.Relationship{
				From:    id.(string),
				To:      targetID.(string),
				Type:    relType.(string),
				Details: details.(string),
			})
		}
	}

	nodeList := make([]models.Person, 0, len(nodes))
	for _, node := range nodes {
		nodeList = append(nodeList, node)
	}

	return models.Graph{Nodes: nodeList, Edges: edges}, nil
}

func GetPersons(ctx context.Context, driver neo4j.DriverWithContext) ([]models.Person, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:Person) 
         RETURN p.id, p.name, p.occupation, p.image_url, p.twitter, p.description`,
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query persons: %w", err)
	}

	var persons []models.Person
	for result.Next(ctx) {
		record := result.Record()
		id, _ := record.Get("p.id")
		name, _ := record.Get("p.name")
		occupation, _ := record.Get("p.occupation")
		imageURL, _ := record.Get("p.image_url")
		twitter, _ := record.Get("p.twitter")
		description, _ := record.Get("p.description")

		persons = append(persons, models.Person{
			ID:          id.(string),
			Name:        name.(string),
			Occupation:  occupation.(string),
			ImageURL:    imageURL.(string),
			Twitter:     twitter.(string),
			Description: description.(string),
		})
	}

	return persons, nil
}

func AddUser(ctx context.Context, driver neo4j.DriverWithContext, user models.User) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (u:User) WHERE u.login = $login OR u.email = $email
         RETURN u`,
		map[string]interface{}{
			"login": user.Login,
			"email": user.Email,
		})
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if result.Next(ctx) {
		return ErrUserExists
	}

	_, err = session.Run(ctx,
		`CREATE (u:User {id: $id, login: $login, email: $email, password: $password})`,
		map[string]interface{}{
			"id":       user.ID,
			"login":    user.Login,
			"email":    user.Email,
			"password": user.Password,
		})
	if err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}
	return nil
}

func GetUserByLogin(ctx context.Context, driver neo4j.DriverWithContext, login string) (models.User, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (u:User {login: $login})
         RETURN u.id, u.login, u.email, u.password`,
		map[string]interface{}{"login": login})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to query user: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		id, _ := record.Get("u.id")
		login, _ := record.Get("u.login")
		email, _ := record.Get("u.email")
		password, _ := record.Get("u.password")

		return models.User{
			ID:       id.(string),
			Login:    login.(string),
			Email:    email.(string),
			Password: password.(string),
		}, nil
	}

	return models.User{}, fmt.Errorf("user not found")
}

func GetUserByID(ctx context.Context, driver neo4j.DriverWithContext, id string) (models.User, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (u:User {id: $id})
         RETURN u.id, u.login, u.email, u.password`,
		map[string]interface{}{"id": id})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to query user: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		id, _ := record.Get("u.id")
		login, _ := record.Get("u.login")
		email, _ := record.Get("u.email")
		password, _ := record.Get("u.password")

		return models.User{
			ID:       id.(string),
			Login:    login.(string),
			Email:    email.(string),
			Password: password.(string),
		}, nil
	}

	return models.User{}, fmt.Errorf("user not found")
}

func CreateSession(ctx context.Context, driver neo4j.DriverWithContext, session models.Session) error {
	neo4jSession := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close(ctx)

	_, err := neo4jSession.Run(ctx,
		`CREATE (s:Session {id: $id, userId: $userId, expiresAt: $expiresAt})`,
		map[string]interface{}{
			"id":        session.ID,
			"userId":    session.UserID,
			"expiresAt": session.ExpiresAt,
		})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func GetSession(ctx context.Context, driver neo4j.DriverWithContext, sessionID string) (models.Session, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (s:Session {id: $id})
         RETURN s.id, s.userId, s.expiresAt`,
		map[string]interface{}{"id": sessionID})
	if err != nil {
		return models.Session{}, fmt.Errorf("failed to query session: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		id, _ := record.Get("s.id")
		userID, _ := record.Get("s.userId")
		expiresAt, _ := record.Get("s.expiresAt")

		return models.Session{
			ID:        id.(string),
			UserID:    userID.(string),
			ExpiresAt: expiresAt.(int64),
		}, nil
	}

	return models.Session{}, ErrNoSuchSession
}

func DeleteSession(ctx context.Context, driver neo4j.DriverWithContext, sessionID string) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		`MATCH (s:Session {id: $id})
         DELETE s`,
		map[string]interface{}{"id": sessionID})
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
