package person

import "errors"

var (
	ErrMissingRelationshipEnd  = errors.New("relationship source and target ids are required")
	ErrMissingRelationshipType = errors.New("relationship type is required")
	ErrSelfRelationship        = errors.New("person cannot be in a relationship with themselves")
)

// Relationship is directed and unique per (SourceID, TargetID, Type);
// adding it again replaces its Details.
type Relationship struct {
	SourceID string
	TargetID string
	Type     string
	Details  string
}

func (r Relationship) Validate() error {
	if isBlank(r.SourceID) || isBlank(r.TargetID) {
		return ErrMissingRelationshipEnd
	}
	if isBlank(r.Type) {
		return ErrMissingRelationshipType
	}
	if r.SourceID == r.TargetID {
		return ErrSelfRelationship
	}
	return nil
}

type Graph struct {
	Persons       []Person
	Relationships []Relationship
}
