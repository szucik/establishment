package database

import (
	"context"

	"establishment/v1/establishment/models"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ConnectToNeo4j(ctx context.Context) (neo4j.DriverWithContext, error) {
	uri := "neo4j://localhost:7687" // Adres bazy Neo4j
	username := "neo4j"
	password := "secretgraph"
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))

	if err != nil {
		return nil, err
	}
	// Weryfikacja połączenia z użyciem kontekstu
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, err
	}
	return driver, nil
}

func AddPerson(ctx context.Context, driver neo4j.DriverWithContext, person models.Person) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer func() {
		if err := session.Close(ctx); err != nil {
			// Logowanie błędu zamknięcia sesji
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (p:Person {
				id: $id,
				name: $name,
				occupation: $occupation,
				image_url: $image_url,
				twitter: $twitter,
				description: $description
			})
			RETURN p
		`

		params := map[string]interface{}{
			"id":          person.ID,
			"name":        person.Name,
			"occupation":  person.Occupation,
			"image_url":   person.ImageURL,
			"twitter":     person.Twitter,
			"description": person.Description,
		}
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		_, err = result.Consume(ctx)
		return nil, err
	})

	return err
}

func AddRelationship(ctx context.Context, driver neo4j.DriverWithContext, rel models.Relationship) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer func() {
		if err := session.Close(ctx); err != nil {
			// Logowanie błędu zamknięcia sesji
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (source:Person {id: $source_id}), (target:Person {id: $target_id})
			MERGE (source)-[r:RELATION {type: $type, details: $details}]->(target)
			RETURN source, target
		`
		params := map[string]interface{}{
			"source_id": rel.SourceID,
			"target_id": rel.TargetID,
			"type":      rel.Type,
			"details":   rel.Details,
		}
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		_, err = result.Consume(ctx)
		return nil, err
	})

	return err
}

func GetPersons(ctx context.Context, driver neo4j.DriverWithContext) ([]models.Person, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer func() {
		if err := session.Close(ctx); err != nil {
			// Logowanie błędu zamknięcia sesji
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (p:Person) RETURN p.id, p.name, p.occupation`
		records, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		var persons []models.Person
		for records.Next(ctx) {
			record := records.Record()
			person := models.Person{
				ID:         record.Values[0].(string),
				Name:       record.Values[1].(string),
				Occupation: record.Values[2].(string),
			}
			persons = append(persons, person)
		}
		return persons, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]models.Person), nil
}

func GetGraph(ctx context.Context, driver neo4j.DriverWithContext) (models.Graph, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer func() {
		if err := session.Close(ctx); err != nil {
			// Logowanie błędu zamknięcia sesji
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
            MATCH (p:Person)-[r:RELATION]->(p2:Person)
            RETURN 
                p.id, p.name, p.occupation, p.image_url, p.twitter, p.description,
                r.type, r.details,
                p2.id, p2.name, p2.occupation, p2.image_url, p2.twitter, p2.description
        `
		records, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		nodes := make(map[string]models.Node)
		var edges []models.Edge

		for records.Next(ctx) {
			record := records.Record()

			// Dane źródła (p)
			sourceID := record.Values[0].(string)
			sourceName := record.Values[1].(string)
			sourceOccupation := record.Values[2].(string)

			var sourceImageURL, sourceTwitter, sourceDescription string

			if record.Values[3] != nil {
				sourceImageURL = record.Values[3].(string)
			}
			if record.Values[4] != nil {
				sourceTwitter = record.Values[4].(string)
			}
			if record.Values[5] != nil {
				sourceDescription = record.Values[5].(string)
			}

			// Relacja
			relType := record.Values[6].(string)
			relDetails := record.Values[7].(string)

			// Dane celu (p2)
			targetID := record.Values[8].(string)
			targetName := record.Values[9].(string)
			targetOccupation := record.Values[10].(string)

			var targetImageURL, targetTwitter, targetDescription string

			if record.Values[11] != nil {
				targetImageURL = record.Values[11].(string)
			}
			if record.Values[12] != nil {
				targetTwitter = record.Values[12].(string)
			}
			if record.Values[13] != nil {
				targetDescription = record.Values[13].(string)
			}

			// Dodajemy węzły do mapy (unikalne)
			nodes[sourceID] = models.Node{
				ID:          sourceID,
				Name:        sourceName,
				Occupation:  sourceOccupation,
				ImageURL:    sourceImageURL,
				Twitter:     sourceTwitter,
				Description: sourceDescription,
			}
			nodes[targetID] = models.Node{
				ID:          targetID,
				Name:        targetName,
				Occupation:  targetOccupation,
				ImageURL:    targetImageURL,
				Twitter:     targetTwitter,
				Description: targetDescription,
			}

			// Dodajemy krawędź
			edges = append(edges, models.Edge{
				Source:  sourceID,
				Target:  targetID,
				Type:    relType,
				Details: relDetails,
			})
		}

		graph := models.Graph{}
		for _, node := range nodes {
			graph.Nodes = append(graph.Nodes, node)
		}
		graph.Edges = edges
		return graph, nil
	})
	if err != nil {
		return models.Graph{}, err
	}
	return result.(models.Graph), nil
}
