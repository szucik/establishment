package database

import (
	"context"
	"errors"

	"establishment/v1/establishment/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ErrNoSuchPerson = errors.New("no such person")

func ConnectToNeo4j(ctx context.Context) (neo4j.DriverWithContext, error) {
	uri := "neo4j://localhost:7687"
	username := "neo4j"
	password := "secretgraph"
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, err
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, err
	}
	return driver, nil
}

func AddPerson(ctx context.Context, driver neo4j.DriverWithContext, person models.Person) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer func() {
		if err := session.Close(ctx); err != nil {
			// Log session close error
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (p:Person {
				id: $id,
				name: $name,
				occupation: $occupation,
				party: $party,
				sb_status: $sb_status,
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
			"party":       person.Party,
			"sb_status":   person.SBStatus,
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
			// Log session close error
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
			// Log session close error
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (p:Person) RETURN p.id, p.name, p.occupation, p.party, p.sb_status`
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
			if record.Values[3] != nil {
				person.Party = record.Values[3].(string)
			}
			if record.Values[4] != nil {
				person.SBStatus = record.Values[4].(string)
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

func GetPerson(ctx context.Context, driver neo4j.DriverWithContext, id string) (models.Person, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer func() {
		if err := session.Close(ctx); err != nil {
			// Log session close error
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (p:Person {id: $id})
			RETURN p.id, p.name, p.occupation, p.party, p.sb_status, p.image_url, p.twitter, p.description
		`
		params := map[string]interface{}{
			"id": id,
		}
		records, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		if records.Next(ctx) {
			record := records.Record()
			person := models.Person{
				ID:         record.Values[0].(string),
				Name:       record.Values[1].(string),
				Occupation: record.Values[2].(string),
			}
			if record.Values[3] != nil {
				person.Party = record.Values[3].(string)
			}
			if record.Values[4] != nil {
				person.SBStatus = record.Values[4].(string)
			}
			if record.Values[5] != nil {
				person.ImageURL = record.Values[5].(string)
			}
			if record.Values[6] != nil {
				person.Twitter = record.Values[6].(string)
			}
			if record.Values[7] != nil {
				person.Description = record.Values[7].(string)
			}
			return person, nil
		}
		return nil, ErrNoSuchPerson
	})
	if err != nil {
		return models.Person{}, err
	}
	if result == nil {
		return models.Person{}, ErrNoSuchPerson
	}
	return result.(models.Person), nil
}

func GetGraph(ctx context.Context, driver neo4j.DriverWithContext) (models.Graph, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer func() {
		if err := session.Close(ctx); err != nil {
			// Log session close error
		}
	}()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
            MATCH (p:Person)-[r:RELATION]->(p2:Person)
            RETURN 
                p.id, p.name, p.occupation, p.party, p.sb_status, p.image_url, p.twitter, p.description,
                r.type, r.details,
                p2.id, p2.name, p2.occupation, p2.party, p2.sb_status, p2.image_url, p2.twitter, p2.description
        `
		records, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		nodes := make(map[string]models.Node)
		var edges []models.Edge

		for records.Next(ctx) {
			record := records.Record()

			sourceID := record.Values[0].(string)
			sourceName := record.Values[1].(string)
			sourceOccupation := record.Values[2].(string)
			var sourceParty, sourceSBStatus, sourceImageURL, sourceTwitter, sourceDescription string
			if record.Values[3] != nil {
				sourceParty = record.Values[3].(string)
			}
			if record.Values[4] != nil {
				sourceSBStatus = record.Values[4].(string)
			}
			if record.Values[5] != nil {
				sourceImageURL = record.Values[5].(string)
			}
			if record.Values[6] != nil {
				sourceTwitter = record.Values[6].(string)
			}
			if record.Values[7] != nil {
				sourceDescription = record.Values[7].(string)
			}

			relType := record.Values[8].(string)
			relDetails := record.Values[9].(string)

			targetID := record.Values[10].(string)
			targetName := record.Values[11].(string)
			targetOccupation := record.Values[12].(string)
			var targetParty, targetSBStatus, targetImageURL, targetTwitter, targetDescription string
			if record.Values[13] != nil {
				targetParty = record.Values[13].(string)
			}
			if record.Values[14] != nil {
				targetSBStatus = record.Values[14].(string)
			}
			if record.Values[15] != nil {
				targetImageURL = record.Values[15].(string)
			}
			if record.Values[16] != nil {
				targetTwitter = record.Values[16].(string)
			}
			if record.Values[17] != nil {
				targetDescription = record.Values[17].(string)
			}

			nodes[sourceID] = models.Node{
				ID:          sourceID,
				Name:        sourceName,
				Occupation:  sourceOccupation,
				Party:       sourceParty,
				SBStatus:    sourceSBStatus,
				ImageURL:    sourceImageURL,
				Twitter:     sourceTwitter,
				Description: sourceDescription,
			}
			nodes[targetID] = models.Node{
				ID:          targetID,
				Name:        targetName,
				Occupation:  targetOccupation,
				Party:       targetParty,
				SBStatus:    targetSBStatus,
				ImageURL:    targetImageURL,
				Twitter:     targetTwitter,
				Description: targetDescription,
			}

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
