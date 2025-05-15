package models

type Person struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Occupation  string `json:"occupation"`            // "Polityk" lub "Dziennikarz"
	ImageURL    string `json:"image_url,omitempty"`   // URL zdjęcia
	Twitter     string `json:"twitter,omitempty"`     // Konto na X
	Description string `json:"description,omitempty"` // Opis postaci
}

// Relationship reprezentuje powiązanie między osobami
type Relationship struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Type     string `json:"type"`    // "FAMILY" lub "COLLEAGUE"
	Details  string `json:"details"` // np. "małżeństwo" lub "praca w komisji"
}

// Graph reprezentuje strukturę grafu dla wizualizacji
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Node reprezentuje węzeł w grafie
type Node struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Occupation  string `json:"occupation"`
	ImageURL    string `json:"image_url,omitempty"`
	Twitter     string `json:"twitter,omitempty"`
	Description string `json:"description,omitempty"`
}

// Edge reprezentuje krawędź w grafie
type Edge struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Type    string `json:"type"`
	Details string `json:"details"`
}
