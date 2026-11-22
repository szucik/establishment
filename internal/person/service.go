package person

import (
	"context"
	"fmt"
)

type Repository interface {
	AddPerson(ctx context.Context, p Person) error
	FindPersonByID(ctx context.Context, id string) (Person, error)
	FindAllPersons(ctx context.Context) ([]Person, error)
	SaveRelationship(ctx context.Context, r Relationship) error
	FindAllRelationships(ctx context.Context) ([]Relationship, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) AddPerson(ctx context.Context, p Person) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return s.repository.AddPerson(ctx, p)
}

func (s *Service) Person(ctx context.Context, id string) (Person, error) {
	return s.repository.FindPersonByID(ctx, id)
}

func (s *Service) Persons(ctx context.Context) ([]Person, error) {
	return s.repository.FindAllPersons(ctx)
}

func (s *Service) AddRelationship(ctx context.Context, r Relationship) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if err := s.ensurePersonsExist(ctx, r.SourceID, r.TargetID); err != nil {
		return err
	}
	return s.repository.SaveRelationship(ctx, r)
}

func (s *Service) Graph(ctx context.Context) (Graph, error) {
	persons, err := s.repository.FindAllPersons(ctx)
	if err != nil {
		return Graph{}, err
	}
	relationships, err := s.repository.FindAllRelationships(ctx)
	if err != nil {
		return Graph{}, err
	}
	return Graph{Persons: persons, Relationships: relationships}, nil
}

func (s *Service) ensurePersonsExist(ctx context.Context, ids ...string) error {
	for _, id := range ids {
		if _, err := s.repository.FindPersonByID(ctx, id); err != nil {
			return fmt.Errorf("relationship end %q: %w", id, err)
		}
	}
	return nil
}
