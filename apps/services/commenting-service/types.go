// Package main provides the commenting-service for entity-level comments,
// thread support, author tracking, and project-wide comment feeds.
package main

import (
	"time"
)

// EntityType represents the type of entity a comment is attached to.
type EntityType string

const (
	EntityTypeClass      EntityType = "class"
	EntityTypeProperty   EntityType = "property"
	EntityTypeIndividual EntityType = "individual"
)

// Comment represents a comment attached to an ontology entity.
type Comment struct {
	ID              string     `json:"id"`
	OntologyID      string     `json:"ontology_id"`
	EntityID        string     `json:"entity_id"`
	EntityType      EntityType `json:"entity_type"`
	AuthorID        string     `json:"author_id"`
	Body            string     `json:"body"`
	ParentCommentID *string    `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Metadata        *string    `json:"metadata,omitempty"` // JSONB free-form metadata
	Replies         []Comment  `json:"replies,omitempty"`  // Threaded replies (populated on list/get)
}

// CreateCommentRequest is the JSON body for creating a comment.
type CreateCommentRequest struct {
	EntityID   string  `json:"entity_id"`
	EntityType string  `json:"entity_type"`
	Body       string  `json:"body"`
	ParentID   *string `json:"parent_id,omitempty"`
	Metadata   *string `json:"metadata,omitempty"`
}

// UpdateCommentRequest is the JSON body for updating a comment.
type UpdateCommentRequest struct {
	Body     string  `json:"body"`
	Metadata *string `json:"metadata,omitempty"`
}

// CommentResponse is the API response for a single comment.
type CommentResponse struct {
	ID              string            `json:"id"`
	OntologyID      string            `json:"ontology_id"`
	EntityID        string            `json:"entity_id"`
	EntityType      EntityType        `json:"entity_type"`
	AuthorID        string            `json:"author_id"`
	Body            string            `json:"body"`
	ParentCommentID *string           `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Replies         []CommentResponse `json:"replies,omitempty"`
}

// CommentListResponse is the paginated API response for listing comments.
type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// CommentFeedResponse is the project-wide comment feed response.
type CommentFeedResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int               `json:"total"`
	Since    *time.Time        `json:"since,omitempty"`
}

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// PaginationParams holds pagination query parameters.
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// DefaultPageSize is the default number of items per page.
const DefaultPageSize = 20

// MaxPageSize is the maximum allowed page size.
const MaxPageSize = 100
