package models

// PaginationMeta holds pagination metadata for list endpoints.
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse wraps list endpoint data with pagination metadata.
type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// ErrorDetail represents a single error code and message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps an error in the standard API error format.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// OntologySummary is a compact representation of an ontology for list endpoints.
type OntologySummary struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	ClassCount  int    `json:"class_count"`
	PropCount   int    `json:"property_count"`
	IndivCount  int    `json:"individual_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// OntologyDetail is a full ontology representation with detail counts.
type OntologyDetail struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	ClassCount  int    `json:"class_count"`
	PropCount   int    `json:"property_count"`
	IndivCount  int    `json:"individual_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ClassSummary is a compact representation of an OWL class.
type ClassSummary struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Comment    string `json:"comment,omitempty"`
	ParentID   string `json:"parent_id,omitempty"`
	ChildCount int    `json:"child_count"`
}

// ClassDetail is a full class representation with hierarchy breadcrumb.
type ClassDetail struct {
	ID            string   `json:"id"`
	Label         string   `json:"label"`
	Comment       string   `json:"comment,omitempty"`
	IRI           string   `json:"iri"`
	ParentIDs     []string `json:"parent_ids"`
	Breadcrumb    []string `json:"breadcrumb"`
	PropertyCount int      `json:"property_count"`
	IndivCount    int      `json:"individual_count"`
}

// PropertySummary is a compact representation of an OWL property.
type PropertySummary struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Type         string `json:"type"` // "object" or "datatype"
	Domain       string `json:"domain,omitempty"`
	Range        string `json:"range,omitempty"`
	IsFunctional bool   `json:"is_functional,omitempty"`
}

// IndividualSummary is a compact representation of an ontology individual.
type IndividualSummary struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	ClassID       string `json:"class_id"`
	ClassName     string `json:"class_name,omitempty"`
	PropertyCount int    `json:"property_count"`
}
