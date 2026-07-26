package models

// NlToOwlRequest is the request body for NL→OWL generation.
type NlToOwlRequest struct {
	Text  string `json:"text" binding:"required"`
	Model string `json:"model,omitempty"`
}

// SequenceStep represents a single operation in an ontology sequence.
type SequenceStep struct {
	Operation    string            `json:"operation"`
	EntityID     string            `json:"entity_id"`
	Label        string            `json:"label"`
	ParentID     string            `json:"parent_id,omitempty"`
	Domain       string            `json:"domain,omitempty"`
	Range        string            `json:"range,omitempty"`
	PropertyType string            `json:"property_type,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// SequencePreview is the response returned by NL→OWL generation and
// document extraction. It shows the proposed ontology changes for user review.
type SequencePreview struct {
	OntologyID string         `json:"ontology_id"`
	Steps      []SequenceStep `json:"steps"`
	TotalSteps int            `json:"total_steps"`
	Warnings   []string       `json:"warnings,omitempty"`
}

// SuggestClassesRequest requests AI class completion suggestions.
type SuggestClassesRequest struct {
	ClassID     string `json:"class_id,omitempty"`
	Model       string `json:"model,omitempty"`
	ContextSize int    `json:"context_size,omitempty"` // depth of context to include
}

// SuggestPropertiesRequest requests AI property suggestions for a class.
type SuggestPropertiesRequest struct {
	ClassID string `json:"class_id" binding:"required"`
	Model   string `json:"model,omitempty"`
}

// SuggestRelationshipsRequest requests AI relationship suggestions between classes.
type SuggestRelationshipsRequest struct {
	SourceClassID string `json:"source_class_id" binding:"required"`
	TargetClassID string `json:"target_class_id,omitempty"`
	Model         string `json:"model,omitempty"`
}

// AISuggestion is a single AI-generated suggestion with confidence score.
type AISuggestion struct {
	EntityID            string  `json:"entity_id"`
	Label               string  `json:"label"`
	RelationshipToFocus string  `json:"relationship_to_focus,omitempty"`
	Confidence          float64 `json:"confidence"`
	Rationale           string  `json:"rationale,omitempty"`
	PropertyType        string  `json:"property_type,omitempty"` // "dataProperty" or "objectProperty"
	Domain              string  `json:"domain,omitempty"`
	Range               string  `json:"range,omitempty"`
}

// AISuggestionsResponse wraps a list of AI suggestions.
type AISuggestionsResponse struct {
	Suggestions      []AISuggestion `json:"suggestions"`
	TotalSuggestions int            `json:"total_suggestions"`
	OntologyID       string         `json:"ontology_id"`
}

// RefinementRequest requests iterative refinement of a generated sequence.
type RefinementRequest struct {
	SequenceID   string             `json:"sequence_id" binding:"required"`
	FeedbackText string             `json:"feedback_text" binding:"required"`
	Changes      []RefinementChange `json:"changes,omitempty"`
	Model        string             `json:"model,omitempty"`
}

// RefinementChange represents a specific change requested by the user.
type RefinementChange struct {
	Operation    string `json:"operation"`
	TargetID     string `json:"target_id"`
	Modification string `json:"modification,omitempty"`
}

// RefinementResponse is the response from a refinement request.
type RefinementResponse struct {
	SequenceID    string                   `json:"sequence_id"`
	Changes       []RefinementChangeResult `json:"changes"`
	Summary       string                   `json:"summary"`
	Iteration     int                      `json:"iteration"`
	OntologyID    string                   `json:"ontology_id"`
	MaxIterations int                      `json:"max_iterations"`
}

// RefinementChangeResult describes a single change in the refinement result.
type RefinementChangeResult struct {
	Operation    string `json:"operation"`
	TargetID     string `json:"target_id"`
	Modification string `json:"modification"`
	Rationale    string `json:"rationale"`
}

// SuggestionItem
// SequenceStepHistory holds conversation history for refinement iterations.
type SequenceStepHistory struct {
	SequenceID   string          `json:"sequence_id"`
	OntologyID   string          `json:"ontology_id"`
	OriginalText string          `json:"original_text"`
	Steps        []SequenceStep  `json:"steps"`
	FeedbackLog  []FeedbackEntry `json:"feedback_log"`
	Iteration    int             `json:"iteration"`
}

// FeedbackEntry records a single refinement iteration.
type FeedbackEntry struct {
	Feedback  string                   `json:"feedback"`
	Changes   []RefinementChangeResult `json:"changes"`
	Summary   string                   `json:"summary"`
	Iteration int                      `json:"iteration"`
}
