package person_test

import (
	"context"
	"errors"
	"testing"

	"establishment/v1/internal/person"
)

type fakeRepository struct {
	persons       map[string]person.Person
	relationships []person.Relationship
}

func newFakeRepository(persons ...person.Person) *fakeRepository {
	repository := &fakeRepository{persons: map[string]person.Person{}}
	for _, p := range persons {
		repository.persons[p.ID] = p
	}
	return repository
}

func (f *fakeRepository) AddPerson(_ context.Context, p person.Person) error {
	if _, exists := f.persons[p.ID]; exists {
		return person.ErrAlreadyExists
	}
	f.persons[p.ID] = p
	return nil
}

func (f *fakeRepository) FindPersonByID(_ context.Context, id string) (person.Person, error) {
	p, ok := f.persons[id]
	if !ok {
		return person.Person{}, person.ErrNotFound
	}
	return p, nil
}

func (f *fakeRepository) FindAllPersons(context.Context) ([]person.Person, error) {
	persons := []person.Person{}
	for _, p := range f.persons {
		persons = append(persons, p)
	}
	return persons, nil
}

func (f *fakeRepository) SaveRelationship(_ context.Context, r person.Relationship) error {
	f.relationships = append(f.relationships, r)
	return nil
}

func (f *fakeRepository) FindAllRelationships(context.Context) ([]person.Relationship, error) {
	return f.relationships, nil
}

var (
	jan  = person.Person{ID: "jan", Name: "Jan"}
	anna = person.Person{ID: "anna", Name: "Anna"}
)

func TestServiceAddPerson_InvalidPerson_ReturnsValidationError(t *testing.T) {
	service := person.NewService(newFakeRepository())

	err := service.AddPerson(context.Background(), person.Person{ID: "1"})

	if !errors.Is(err, person.ErrMissingName) {
		t.Errorf("AddPerson() = %v, want %v", err, person.ErrMissingName)
	}
}

func TestServiceAddPerson_ValidPerson_PersonIsRetrievable(t *testing.T) {
	service := person.NewService(newFakeRepository())

	if err := service.AddPerson(context.Background(), jan); err != nil {
		t.Fatalf("AddPerson() = %v", err)
	}

	if got, err := service.Person(context.Background(), jan.ID); err != nil || got != jan {
		t.Errorf("Person() = %+v, %v; want %+v", got, err, jan)
	}
}

func TestServiceAddRelationship_MissingPerson_ReturnsNotFound(t *testing.T) {
	service := person.NewService(newFakeRepository(jan))

	err := service.AddRelationship(context.Background(), person.Relationship{SourceID: jan.ID, TargetID: "ghost", Type: "partner"})

	if !errors.Is(err, person.ErrNotFound) {
		t.Errorf("AddRelationship() = %v, want %v", err, person.ErrNotFound)
	}
}

func TestServiceAddRelationship_SelfReference_ReturnsValidationError(t *testing.T) {
	service := person.NewService(newFakeRepository(jan))

	err := service.AddRelationship(context.Background(), person.Relationship{SourceID: jan.ID, TargetID: jan.ID, Type: "partner"})

	if !errors.Is(err, person.ErrSelfRelationship) {
		t.Errorf("AddRelationship() = %v, want %v", err, person.ErrSelfRelationship)
	}
}

func TestServiceGraph_PersonsWithRelationship_ReturnsBoth(t *testing.T) {
	service := person.NewService(newFakeRepository(jan, anna))
	relationship := person.Relationship{SourceID: jan.ID, TargetID: anna.ID, Type: "partner"}
	if err := service.AddRelationship(context.Background(), relationship); err != nil {
		t.Fatalf("AddRelationship() = %v", err)
	}

	graph, err := service.Graph(context.Background())

	if err != nil || len(graph.Persons) != 2 || len(graph.Relationships) != 1 || graph.Relationships[0] != relationship {
		t.Errorf("Graph() = %+v, %v; want 2 persons and %+v", graph, err, relationship)
	}
}
