package neo4jstore

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"establishment/v1/internal/person"
)

const returnPerson = `
	RETURN p.id AS id, COALESCE(p.name, '') AS name, COALESCE(p.occupation, '') AS occupation,
	       COALESCE(p.image_url, '') AS image_url, COALESCE(p.twitter, '') AS twitter,
	       COALESCE(p.description, '') AS description`

func (s *Store) AddPerson(ctx context.Context, p person.Person) error {
	_, err := s.write(ctx,
		`CREATE (:Person {id: $id, name: $name, occupation: $occupation, image_url: $image_url,
		                  twitter: $twitter, description: $description})`,
		map[string]any{
			"id": p.ID, "name": p.Name, "occupation": p.Occupation,
			"image_url": p.ImageURL, "twitter": p.Twitter, "description": p.Description,
		})
	if isConstraintViolation(err) {
		return person.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("creating person %q: %w", p.ID, err)
	}
	return nil
}

func (s *Store) FindPersonByID(ctx context.Context, id string) (person.Person, error) {
	records, err := s.read(ctx, `MATCH (p:Person {id: $id})`+returnPerson, map[string]any{"id": id})
	if err != nil {
		return person.Person{}, fmt.Errorf("querying person %q: %w", id, err)
	}
	if len(records) == 0 {
		return person.Person{}, person.ErrNotFound
	}
	return personFrom(records[0])
}

func (s *Store) FindAllPersons(ctx context.Context) ([]person.Person, error) {
	records, err := s.read(ctx, `MATCH (p:Person)`+returnPerson+` ORDER BY name`, nil)
	if err != nil {
		return nil, fmt.Errorf("querying persons: %w", err)
	}
	persons := make([]person.Person, 0, len(records))
	for _, record := range records {
		p, mapErr := personFrom(record)
		if mapErr != nil {
			return nil, mapErr
		}
		persons = append(persons, p)
	}
	return persons, nil
}

func (s *Store) SaveRelationship(ctx context.Context, r person.Relationship) error {
	records, err := s.write(ctx,
		`MATCH (a:Person {id: $source_id}), (b:Person {id: $target_id})
		 MERGE (a)-[rel:RELATIONSHIP {type: $type}]->(b)
		 SET rel.details = $details
		 RETURN rel.type AS type`,
		map[string]any{"source_id": r.SourceID, "target_id": r.TargetID, "type": r.Type, "details": r.Details})
	if err != nil {
		return fmt.Errorf("saving relationship %s->%s: %w", r.SourceID, r.TargetID, err)
	}
	if len(records) == 0 {
		return person.ErrNotFound
	}
	return nil
}

func (s *Store) FindAllRelationships(ctx context.Context) ([]person.Relationship, error) {
	records, err := s.read(ctx,
		`MATCH (a:Person)-[rel:RELATIONSHIP]->(b:Person)
		 RETURN a.id AS source_id, b.id AS target_id,
		        COALESCE(rel.type, '') AS type, COALESCE(rel.details, '') AS details`, nil)
	if err != nil {
		return nil, fmt.Errorf("querying relationships: %w", err)
	}
	relationships := make([]person.Relationship, 0, len(records))
	for _, record := range records {
		r, mapErr := relationshipFrom(record)
		if mapErr != nil {
			return nil, mapErr
		}
		relationships = append(relationships, r)
	}
	return relationships, nil
}

func personFrom(record *neo4j.Record) (person.Person, error) {
	reader := &recordReader{record: record}
	p := person.Person{
		ID:          reader.string("id"),
		Name:        reader.string("name"),
		Occupation:  reader.string("occupation"),
		ImageURL:    reader.string("image_url"),
		Twitter:     reader.string("twitter"),
		Description: reader.string("description"),
	}
	return p, reader.err
}

func relationshipFrom(record *neo4j.Record) (person.Relationship, error) {
	reader := &recordReader{record: record}
	r := person.Relationship{
		SourceID: reader.string("source_id"),
		TargetID: reader.string("target_id"),
		Type:     reader.string("type"),
		Details:  reader.string("details"),
	}
	return r, reader.err
}
