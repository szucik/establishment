package person_test

import (
	"errors"
	"testing"

	"establishment/v1/internal/person"
)

func TestPersonValidate_InvalidPerson_ReturnsMatchingError(t *testing.T) {
	tests := map[string]struct {
		person person.Person
		want   error
	}{
		"missing id":   {person.Person{Name: "Jan"}, person.ErrMissingID},
		"blank id":     {person.Person{ID: "  ", Name: "Jan"}, person.ErrMissingID},
		"missing name": {person.Person{ID: "1"}, person.ErrMissingName},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tc.person.Validate(); !errors.Is(err, tc.want) {
				t.Errorf("Validate() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestPersonValidate_IDAndName_ReturnsNil(t *testing.T) {
	p := person.Person{ID: "1", Name: "Jan"}

	if err := p.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestRelationshipValidate_InvalidRelationship_ReturnsMatchingError(t *testing.T) {
	tests := map[string]struct {
		relationship person.Relationship
		want         error
	}{
		"missing source": {person.Relationship{TargetID: "2", Type: "partner"}, person.ErrMissingRelationshipEnd},
		"missing target": {person.Relationship{SourceID: "1", Type: "partner"}, person.ErrMissingRelationshipEnd},
		"missing type":   {person.Relationship{SourceID: "1", TargetID: "2"}, person.ErrMissingRelationshipType},
		"self reference": {person.Relationship{SourceID: "1", TargetID: "1", Type: "partner"}, person.ErrSelfRelationship},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tc.relationship.Validate(); !errors.Is(err, tc.want) {
				t.Errorf("Validate() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestRelationshipValidate_TwoDifferentPersonsWithType_ReturnsNil(t *testing.T) {
	r := person.Relationship{SourceID: "1", TargetID: "2", Type: "partner"}

	if err := r.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}
