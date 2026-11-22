package httpapi

import (
	"net/http"

	"establishment/v1/internal/person"
)

type personJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Occupation  string `json:"occupation"`
	ImageURL    string `json:"image_url"`
	Twitter     string `json:"twitter"`
	Description string `json:"description"`
}

type relationshipJSON struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Type     string `json:"type"`
	Details  string `json:"details"`
}

type graphJSON struct {
	Nodes []personJSON       `json:"nodes"`
	Edges []relationshipJSON `json:"edges"`
}

type personHandlers struct {
	persons personCatalog
}

func (h personHandlers) getPerson(w http.ResponseWriter, r *http.Request) {
	found, err := h.persons.Person(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPersonJSON(found))
}

func (h personHandlers) listPersons(w http.ResponseWriter, r *http.Request) {
	persons, err := h.persons.Persons(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPersonsJSON(persons))
}

func (h personHandlers) getGraph(w http.ResponseWriter, r *http.Request) {
	graph, err := h.persons.Graph(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toGraphJSON(graph))
}

func (h personHandlers) postPerson(w http.ResponseWriter, r *http.Request) {
	var input personJSON
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.persons.AddPerson(r.Context(), input.toPerson()); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h personHandlers) postRelationship(w http.ResponseWriter, r *http.Request) {
	var input relationshipJSON
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.persons.AddRelationship(r.Context(), input.toRelationship()); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (p personJSON) toPerson() person.Person {
	return person.Person{
		ID: p.ID, Name: p.Name, Occupation: p.Occupation,
		ImageURL: p.ImageURL, Twitter: p.Twitter, Description: p.Description,
	}
}

func (r relationshipJSON) toRelationship() person.Relationship {
	return person.Relationship{SourceID: r.SourceID, TargetID: r.TargetID, Type: r.Type, Details: r.Details}
}

func toPersonJSON(p person.Person) personJSON {
	return personJSON{
		ID: p.ID, Name: p.Name, Occupation: p.Occupation,
		ImageURL: p.ImageURL, Twitter: p.Twitter, Description: p.Description,
	}
}

func toPersonsJSON(persons []person.Person) []personJSON {
	result := make([]personJSON, 0, len(persons))
	for _, p := range persons {
		result = append(result, toPersonJSON(p))
	}
	return result
}

func toGraphJSON(graph person.Graph) graphJSON {
	edges := make([]relationshipJSON, 0, len(graph.Relationships))
	for _, r := range graph.Relationships {
		edges = append(edges, relationshipJSON{SourceID: r.SourceID, TargetID: r.TargetID, Type: r.Type, Details: r.Details})
	}
	return graphJSON{Nodes: toPersonsJSON(graph.Persons), Edges: edges}
}
