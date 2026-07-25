// Package main provides PostgreSQL-backed storage for the commenting service.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// CommentStore handles Comment CRUD operations against PostgreSQL.
type CommentStore struct {
	db *sql.DB
}

// NewCommentStore initialises a PostgreSQL connection pool and returns a CommentStore.
// The DSN is read from the DATABASE_URL environment variable.
func NewCommentStore(databaseURL string) (*CommentStore, error) {
	slog.Info("Connecting to PostgreSQL", "database_url", maskDSN(databaseURL))

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &CommentStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("PostgreSQL connection established and schema migrated")
	return store, nil
}

// maskDSN hides the password portion of a DSN for logging.
func maskDSN(dsn string) string {
	// Simple heuristic: hide the password part between ":" and "@"
	runes := []rune(dsn)
	var masked []rune
	passStart := -1
	for i, c := range runes {
		if c == ':' && i > 0 && runes[i-1] != '/' {
			passStart = i + 1
		}
		if c == '@' && passStart >= 0 {
			masked = append(masked, []rune(string(runes[:passStart]))...)
			masked = append(masked, []rune("****")...)
			masked = append(masked, runes[i:]...)
			return string(masked)
		}
	}
	return dsn
}

// migrate creates the comments table if it does not exist.
func (s *CommentStore) migrate(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS comments (
		id            TEXT PRIMARY KEY,
		ontology_id   TEXT NOT NULL,
		entity_id     TEXT NOT NULL,
		entity_type   TEXT NOT NULL CHECK (entity_type IN ('class', 'property', 'individual')),
		author_id     TEXT NOT NULL,
		body          TEXT NOT NULL,
		parent_comment_id TEXT,
		created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		metadata      JSONB DEFAULT '{}'::jsonb,
		FOREIGN KEY (parent_comment_id) REFERENCES comments(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_comments_ontology_entity
		ON comments(ontology_id, entity_id, created_at DESC);

	CREATE INDEX IF NOT EXISTS idx_comments_ontology_feed
		ON comments(ontology_id, created_at DESC);

	CREATE INDEX IF NOT EXISTS idx_comments_parent
		ON comments(parent_comment_id);

	CREATE INDEX IF NOT EXISTS idx_comments_author
		ON comments(author_id, created_at DESC);
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

// CreateComment inserts a new comment and returns it.
func (s *CommentStore) CreateComment(ctx context.Context, ontologyID string, req CreateCommentRequest, authorID string) (*Comment, error) {
	slog.Debug("Creating comment",
		"ontology_id", ontologyID,
		"entity_id", req.EntityID,
		"author_id", authorID,
	)

	now := time.Now().UTC()
	id := fmt.Sprintf("cmt_%d", now.UnixNano())

	meta := "{}"
	if req.Metadata != nil {
		meta = *req.Metadata
		// Validate that metadata is valid JSON
		if !json.Valid([]byte(meta)) {
			return nil, fmt.Errorf("metadata must be valid JSON")
		}
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO comments (id, ontology_id, entity_id, entity_type, author_id, body, parent_comment_id, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
	`,
		id, ontologyID, req.EntityID, req.EntityType, authorID, req.Body, req.ParentID, now, now, meta,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert comment: %w", err)
	}

	slog.Info("Comment created",
		"comment_id", id,
		"ontology_id", ontologyID,
		"entity_id", req.EntityID,
		"author_id", authorID,
	)

	return &Comment{
		ID:              id,
		OntologyID:      ontologyID,
		EntityID:        req.EntityID,
		EntityType:      EntityType(req.EntityType),
		AuthorID:        authorID,
		Body:            req.Body,
		ParentCommentID: req.ParentID,
		CreatedAt:       now,
		UpdatedAt:       now,
		Metadata:        req.Metadata,
	}, nil
}

// GetComment retrieves a single comment by ID.
func (s *CommentStore) GetComment(ctx context.Context, id string) (*Comment, error) {
	slog.Debug("Getting comment", "comment_id", id)

	row := s.db.QueryRowContext(ctx, `
		SELECT id, ontology_id, entity_id, entity_type, author_id, body, parent_comment_id, created_at, updated_at, metadata
		FROM comments WHERE id = $1
	`, id)

	c, err := scanComment(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}
	return c, nil
}

// UpdateComment updates an existing comment's body and/or metadata.
func (s *CommentStore) UpdateComment(ctx context.Context, id string, req UpdateCommentRequest) (*Comment, error) {
	slog.Debug("Updating comment", "comment_id", id)

	meta := "{}"
	if req.Metadata != nil {
		if !json.Valid([]byte(*req.Metadata)) {
			return nil, fmt.Errorf("metadata must be valid JSON")
		}
		meta = *req.Metadata
	}

	slog.Debug("[FIX] UpdateComment with parameterized metadata", "comment_id", id)

	row := s.db.QueryRowContext(ctx, `
		UPDATE comments SET body = $2, metadata = $3::jsonb, updated_at = NOW()
		WHERE id = $1
		RETURNING id, ontology_id, entity_id, entity_type, author_id, body, parent_comment_id, created_at, updated_at, metadata
	`, id, req.Body, meta)

	c, err := scanComment(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	slog.Info("Comment updated", "comment_id", id)
	return c, nil
}

// DeleteComment soft-removes a comment by ID (cascade deletes replies via FK).
func (s *CommentStore) DeleteComment(ctx context.Context, id string) error {
	slog.Debug("Deleting comment", "comment_id", id)

	result, err := s.db.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("comment not found: %s", id)
	}

	slog.Info("Comment deleted", "comment_id", id)
	return nil
}

// ListCommentsByEntity returns paginated comments for a specific entity.
// Top-level comments include their threaded replies (nested under parent_comment_id).
func (s *CommentStore) ListCommentsByEntity(ctx context.Context, ontologyID, entityID string, page, pageSize int) ([]Comment, int, error) {
	slog.Debug("Listing comments by entity",
		"ontology_id", ontologyID,
		"entity_id", entityID,
		"page", page,
		"page_size", pageSize,
	)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		pageSize = DefaultPageSize
	}
	offset := (page - 1) * pageSize

	// Count total
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM comments WHERE ontology_id = $1 AND entity_id = $2`,
		ontologyID, entityID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// Fetch paginated top-level comments (no parent)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, ontology_id, entity_id, entity_type, author_id, body, parent_comment_id, created_at, updated_at, metadata
		FROM comments
		WHERE ontology_id = $1 AND entity_id = $2 AND parent_comment_id IS NULL
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, ontologyID, entityID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, *c)
	}

	// Fetch replies for each top-level comment
	for i, c := range comments {
		replies, err := s.getReplies(ctx, ontologyID, c.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to fetch replies: %w", err)
		}
		// [FIX] Actually attach fetched replies to the comment instead of discarding
		comments[i].Replies = replies
	}

	return comments, total, nil
}

// getRepliesRaw fetches all direct replies to a given parent comment (used by handlers).
func (s *CommentStore) getRepliesRaw(ctx context.Context, ontologyID, parentID string) ([]Comment, error) {
	return s.getReplies(ctx, ontologyID, parentID)
}

// getReplies fetches all direct replies to a given parent comment.
func (s *CommentStore) getReplies(ctx context.Context, ontologyID, parentID string) ([]Comment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, ontology_id, entity_id, entity_type, author_id, body, parent_comment_id, created_at, updated_at, metadata
		FROM comments
		WHERE ontology_id = $1 AND parent_comment_id = $2
		ORDER BY created_at ASC
	`, ontologyID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []Comment
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		replies = append(replies, *c)
	}
	return replies, nil
}

// ListCommentFeed returns a project-wide comment feed, optionally filtered by "since" timestamp.
func (s *CommentStore) ListCommentFeed(ctx context.Context, ontologyID string, since *time.Time, limit int) ([]Comment, int, error) {
	slog.Debug("Listing comment feed",
		"ontology_id", ontologyID,
		"since", since,
		"limit", limit,
	)

	if limit < 1 || limit > MaxPageSize {
		limit = DefaultPageSize
	}

	// Count total
	var total int
	if since != nil {
		err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM comments WHERE ontology_id = $1 AND created_at > $2`,
			ontologyID, *since,
		).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count feed: %w", err)
		}
	} else {
		err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM comments WHERE ontology_id = $1`,
			ontologyID,
		).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count feed: %w", err)
		}
	}

	// Fetch comments ordered by creation time descending
	var rows *sql.Rows
	var err error
	if since != nil {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, ontology_id, entity_id, entity_type, author_id, body, parent_comment_id, created_at, updated_at, metadata
			FROM comments
			WHERE ontology_id = $1 AND created_at > $2
			ORDER BY created_at DESC
			LIMIT $3
		`, ontologyID, *since, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, ontology_id, entity_id, entity_type, author_id, body, parent_comment_id, created_at, updated_at, metadata
			FROM comments
			WHERE ontology_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, ontologyID, limit)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feed: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, *c)
	}

	return comments, total, nil
}

// scanComment scans a single Comment from a row scanner.
func scanComment(row interface{ Scan(dest ...any) error }) (*Comment, error) {
	var c Comment
	var parentID *string
	var metadata sql.NullString

	err := row.Scan(
		&c.ID, &c.OntologyID, &c.EntityID, &c.EntityType,
		&c.AuthorID, &c.Body, &parentID,
		&c.CreatedAt, &c.UpdatedAt, &metadata,
	)
	if err != nil {
		return nil, err
	}

	c.ParentCommentID = parentID
	if metadata.Valid {
		c.Metadata = &metadata.String
	}

	return &c, nil
}

// Close closes the database connection pool.
func (s *CommentStore) Close() error {
	return s.db.Close()
}
