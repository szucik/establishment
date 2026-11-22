package person

import (
	"errors"
	"strings"
)

var (
	ErrMissingID     = errors.New("person id is required")
	ErrMissingName   = errors.New("person name is required")
	ErrNotFound      = errors.New("person not found")
	ErrAlreadyExists = errors.New("person already exists")
)

type Person struct {
	ID          string
	Name        string
	Occupation  string
	ImageURL    string
	Twitter     string
	Description string
}

func (p Person) Validate() error {
	if isBlank(p.ID) {
		return ErrMissingID
	}
	if isBlank(p.Name) {
		return ErrMissingName
	}
	return nil
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}
