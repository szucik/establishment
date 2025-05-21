package models

type Person struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Occupation  string `json:"occupation"`            // "Politician" or "Journalist"
	Party       string `json:"party,omitempty"`       // Name of the political party (required for Politician)
	SBStatus    string `json:"sb_status,omitempty"`   // Status of collaboration with SB or PRL official
	ImageURL    string `json:"image_url,omitempty"`   // Photo URL
	Twitter     string `json:"twitter,omitempty"`     // X account
	Description string `json:"description,omitempty"` // Description of the person
}

// Relationship represents a connection between people
type Relationship struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Type     string `json:"type"`    // "FAMILY" or "COLLEAGUE"
	Details  string `json:"details"` // e.g., "marriage" or "work in a committee"
}

// Graph represents the graph structure for visualization
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node represents a node in the graph
type Node struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Occupation  string `json:"occupation"`
	Party       string `json:"party,omitempty"`     // Name of the political party (required for Politician)
	SBStatus    string `json:"sb_status,omitempty"` // Status of collaboration with SB or PRL official
	ImageURL    string `json:"image_url,omitempty"`
	Twitter     string `json:"twitter,omitempty"`
	Description string `json:"description,omitempty"`
}

// Edge represents an edge in the graph
type Edge struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Type    string `json:"type"`
	Details string `json:"details"`
}
