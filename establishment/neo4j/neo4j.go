package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"establishment/v1/establishment/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	ErrNoSuchPerson        = errors.New("no such person")
	ErrUserExists          = errors.New("user already exists")
	ErrNoSuchSession       = errors.New("no such session")
	ErrInvalidRelationship = errors.New("source and target IDs must be different")
)

func ConnectToNeo4j(ctx context.Context) (neo4j.DriverWithContext, error) {
	neo4jURI := os.Getenv("NEO4J_URI")
	neo4jUser := os.Getenv("NEO4J_USER")
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")

	if neo4jURI == "" {
		neo4jURI = "neo4j://localhost:7687"
		log.Println("Warning: NEO4J_URI not set, using default:", neo4jURI)
	}
	if neo4jUser == "" {
		neo4jUser = "neo4j"
		log.Println("Warning: NEO4J_USER not set, using default:", neo4jUser)
	}
	if neo4jPassword == "" {
		neo4jPassword = "secretgraph"
		log.Println("Warning: NEO4J_PASSWORD not set, using default password")
	}

	log.Printf("Connecting to Neo4j at %s with user %s", neo4jURI, neo4jUser)
	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}
	log.Println("Successfully connected to Neo4j")
	return driver, nil
}

func AddPerson(ctx context.Context, driver neo4j.DriverWithContext, person models.Person) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	log.Printf("Adding person: id=%s, name=%s, occupation=%s", person.ID, person.Name, person.Occupation)

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
		log.Printf("Failed to add person: %v", err)
		return fmt.Errorf("failed to add person: %w", err)
	}
	log.Println("Person added successfully")
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

	// Sprawdź, czy source_id i target_id są różne
	if rel.From == rel.To {
		log.Printf("Invalid relationship: source_id=%s and target_id=%s are the same", rel.From, rel.To)
		return ErrInvalidRelationship
	}

	log.Printf("Verifying persons for relationship: source_id=%s, target_id=%s", rel.From, rel.To)

	result, err := session.Run(ctx,
		`MATCH (a:Person {id: $from}), (b:Person {id: $to})
		 RETURN a.id, b.id`,
		map[string]interface{}{
			"from": rel.From,
			"to":   rel.To,
		})
	if err != nil {
		log.Printf("Failed to verify persons: %v", err)
		return fmt.Errorf("failed to verify persons for relationship: %w", err)
	}
	if !result.Next(ctx) {
		log.Printf("One or both persons not found: source_id=%s, target_id=%s", rel.From, rel.To)
		return fmt.Errorf("one or both persons not found: source_id=%s, target_id=%s", rel.From, rel.To)
	}

	log.Printf("Adding relationship: source_id=%s, target_id=%s, type=%s, details=%s", rel.From, rel.To, rel.Type, rel.Details)

	_, err = session.Run(ctx,
		`MATCH (a:Person {id: $from}), (b:Person {id: $to})
		 MERGE (a)-[r:RELATIONSHIP {type: $type, details: $details}]->(b)
		 RETURN r`,
		map[string]interface{}{
			"from":    rel.From,
			"to":      rel.To,
			"type":    rel.Type,
			"details": rel.Details,
		})
	if err != nil {
		log.Printf("Failed to add relationship: %v", err)
		return fmt.Errorf("failed to add relationship: %w", err)
	}
	log.Printf("Relationship added successfully: %s -> %s (%s)", rel.From, rel.To, rel.Type)
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
		log.Printf("Failed to query graph: %v", err)
		return models.Graph{}, fmt.Errorf("failed to query graph: %w", err)
	}

	nodes := make(map[string]models.Person)
	edges := []models.Relationship{}
	recordCount := 0

	for result.Next(ctx) {
		recordCount++
		log.Printf("Processing record %d: %v", recordCount, result.Record().Values)

		id, ok := result.Record().Get("p.id")
		if !ok || id == nil {
			log.Println("Warning: Missing or nil p.id in graph query result")
			continue
		}

		name, _ := result.Record().Get("p.name")
		occupation, _ := result.Record().Get("p.occupation")
		imageURL, _ := result.Record().Get("p.image_url")
		twitter, _ := result.Record().Get("p.twitter")
		description, _ := result.Record().Get("p.description")

		node := models.Person{
			ID:          id.(string),
			Name:        name.(string),
			Occupation:  occupation.(string),
			ImageURL:    imageURL.(string),
			Twitter:     twitter.(string),
			Description: description.(string),
		}
		nodes[id.(string)] = node

		if targetID, ok := result.Record().Get("target_id"); ok && targetID != nil {
			relType, typeOk := result.Record().Get("r.type")
			details, detailsOk := result.Record().Get("r.details")
			if !typeOk || !detailsOk || relType == nil || details == nil {
				log.Printf("Warning: Missing or nil r.type or r.details for edge: source_id=%s, target_id=%s", id, targetID)
				continue
			}
			if id == targetID {
				log.Printf("Warning: Skipping self-referential edge: source_id=%s, target_id=%s", id, targetID)
				continue
			}
			edge := models.Relationship{
				From:    id.(string),
				To:      targetID.(string),
				Type:    relType.(string),
				Details: details.(string),
			}
			log.Printf("Adding edge: source_id=%s, target_id=%s, type=%s, details=%s", edge.From, edge.To, edge.Type, edge.Details)
			edges = append(edges, edge)
		}
	}

	nodeList := make([]models.Person, 0, len(nodes))
	for _, node := range nodes {
		nodeList = append(nodeList, node)
	}

	validEdges := make([]models.Relationship, 0, len(edges))
	nodeIds := make(map[string]bool)
	for _, node := range nodeList {
		nodeIds[node.ID] = true
	}
	for _, edge := range edges {
		if nodeIds[edge.From] && nodeIds[edge.To] {
			validEdges = append(validEdges, edge)
		} else {
			log.Printf("Skipping invalid edge: source_id=%s, target_id=%s", edge.From, edge.To)
		}
	}

	graph := models.Graph{Nodes: nodeList, Edges: validEdges}
	log.Printf("Returning graph: %d nodes, %d edges", len(graph.Nodes), len(graph.Edges))
	return graph, nil
}

func GetPersons(ctx context.Context, driver neo4j.DriverWithContext) ([]models.Person, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:Person) 
		 RETURN p.id, p.name, p.occupation, p.image_url, p.twitter, p.description`,
		nil)
	if err != nil {
		log.Printf("Failed to query persons: %v", err)
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

	log.Printf("Returning %d persons", len(persons))
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
