// Package main provides HTTP handlers for the commenting service.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// commentHandlers groups HTTP handler methods for comment CRUD operations.
type commentHandlers struct {
	store *CommentStore
	bus   *EventBus
}

// newCommentHandlers creates a new commentHandlers with the given store and event bus.
func newCommentHandlers(store *CommentStore, bus *EventBus) *commentHandlers {
	return &commentHandlers{store: store, bus: bus}
}

// createComment handles POST /api/v1/ontologies/{ontology_id}/comments
func (h *commentHandlers) createComment(w http.ResponseWriter, r *http.Request) {
	ontologyID := r.PathValue("ontology_id")

	// Validate input first, before checking store availability
	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid create comment request body", "error", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST_BODY",
			Message: "Request body must be valid JSON",
		})
		return
	}

	if req.EntityID == "" || req.Body == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "entity_id and body are required",
		})
		return
	}

	if req.EntityType != string(EntityTypeClass) &&
		req.EntityType != string(EntityTypeProperty) &&
		req.EntityType != string(EntityTypeIndividual) {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "entity_type must be one of: class, property, individual",
		})
		return
	}

	// Now check store availability
	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "STORE_UNAVAILABLE",
			Message: "Commenting service has no database connection",
		})
		return
	}

	// Extract author from request context or header
	authorID := getUserID(r)
	if authorID == "" {
		authorID = "anonymous"
	}

	comment, err := h.store.CreateComment(r.Context(), ontologyID, req, authorID)
	if err != nil {
		slog.Error("Failed to create comment", "error", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "CREATE_FAILED",
			Message: "Failed to create comment",
		})
		return
	}

	resp := toCommentResponse(comment, nil)
	slog.Info("Comment created", "comment_id", resp.ID, "ontology_id", ontologyID)

	// Emit event for notification dispatch
	emitCommentCreated(h.bus, comment)

	writeJSON(w, http.StatusCreated, resp)
}

// getComment handles GET /api/v1/comments/{id}
func (h *commentHandlers) getComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "STORE_UNAVAILABLE",
			Message: "Commenting service has no database connection",
		})
		return
	}
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "MISSING_COMMENT_ID",
			Message: "Comment ID is required",
		})
		return
	}

	comment, err := h.store.GetComment(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get comment", "comment_id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "GET_FAILED",
			Message: "Failed to get comment",
		})
		return
	}
	if comment == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "COMMENT_NOT_FOUND",
			Message: "Comment not found: " + id,
		})
		return
	}

	// Fetch replies if this is a top-level comment
	var replies []Comment
	if comment.ParentCommentID == nil {
		var err error
		replies, err = h.store.getRepliesRaw(r.Context(), comment.OntologyID, comment.ID)
		if err != nil {
			slog.Warn("Failed to fetch replies", "comment_id", id, "error", err)
		}
	}

	resp := toCommentResponse(comment, replies)
	writeJSON(w, http.StatusOK, resp)
}

// updateComment handles PUT /api/v1/comments/{id}
func (h *commentHandlers) updateComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Invalid update comment request body", "error", err)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST_BODY",
			Message: "Request body must be valid JSON",
		})
		return
	}

	if req.Body == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "body is required",
		})
		return
	}

	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "STORE_UNAVAILABLE",
			Message: "Commenting service has no database connection",
		})
		return
	}

	comment, err := h.store.UpdateComment(r.Context(), id, req)
	if err != nil {
		slog.Error("Failed to update comment", "comment_id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "UPDATE_FAILED",
			Message: "Failed to update comment",
		})
		return
	}
	if comment == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "COMMENT_NOT_FOUND",
			Message: "Comment not found: " + id,
		})
		return
	}

	resp := toCommentResponse(comment, nil)
	slog.Info("Comment updated", "comment_id", id)

	// Emit event for notification dispatch
	emitCommentUpdated(h.bus, comment)

	writeJSON(w, http.StatusOK, resp)
}

// deleteComment handles DELETE /api/v1/comments/{id}
func (h *commentHandlers) deleteComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "STORE_UNAVAILABLE",
			Message: "Commenting service has no database connection",
		})
		return
	}
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "MISSING_COMMENT_ID",
			Message: "Comment ID is required",
		})
		return
	}

	err := h.store.DeleteComment(r.Context(), id)
	if err != nil {
		slog.Warn("Failed to delete comment", "comment_id", id, "error", err)
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "COMMENT_NOT_FOUND",
			Message: "Comment not found: " + id,
		})
		return
	}

	slog.Info("Comment deleted", "comment_id", id)

	// Emit event for notification dispatch
	emitCommentDeleted(h.bus, id, "")

	w.WriteHeader(http.StatusNoContent)
}

// listCommentsByEntity handles GET /api/v1/ontologies/{ontology_id}/comments
// Query params: entity_id (required), page, page_size
func (h *commentHandlers) listCommentsByEntity(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("entity_id")
	if entityID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "MISSING_ENTITY_ID",
			Message: "entity_id query parameter is required",
		})
		return
	}

	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "STORE_UNAVAILABLE",
			Message: "Commenting service has no database connection",
		})
		return
	}
	ontologyID := r.PathValue("ontology_id")
	if entityID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "MISSING_ENTITY_ID",
			Message: "entity_id query parameter is required",
		})
		return
	}

	page := parseIntParam(r.URL.Query().Get("page"), 1)
	pageSize := parseIntParam(r.URL.Query().Get("page_size"), DefaultPageSize)

	comments, total, err := h.store.ListCommentsByEntity(r.Context(), ontologyID, entityID, page, pageSize)
	if err != nil {
		slog.Error("Failed to list comments",
			"ontology_id", ontologyID,
			"entity_id", entityID,
			"error", err,
		)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "LIST_FAILED",
			Message: "Failed to list comments",
		})
		return
	}

	responses := make([]CommentResponse, 0, len(comments))
	for _, c := range comments {
		// [FIX] Pass embedded replies through to the API response
		responses = append(responses, *toCommentResponse(&c, c.Replies))
	}

	slog.Debug("Listed comments by entity",
		"ontology_id", ontologyID,
		"entity_id", entityID,
		"total", total,
	)

	writeJSON(w, http.StatusOK, CommentListResponse{
		Comments: responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// listCommentFeed handles GET /api/v1/ontologies/{ontology_id}/comment-feed
// Query params: since (RFC3339 timestamp), limit
func (h *commentHandlers) listCommentFeed(w http.ResponseWriter, r *http.Request) {
	// Validate 'since' parameter first (no DB needed)
	var since *time.Time
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		t, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "INVALID_SINCE_PARAM",
				Message: "since parameter must be in RFC3339 format",
			})
			return
		}
		since = &t
	}

	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "STORE_UNAVAILABLE",
			Message: "Commenting service has no database connection",
		})
		return
	}
	ontologyID := r.PathValue("ontology_id")

	limit := parseIntParam(r.URL.Query().Get("limit"), DefaultPageSize)

	comments, total, err := h.store.ListCommentFeed(r.Context(), ontologyID, since, limit)
	if err != nil {
		slog.Error("Failed to list comment feed",
			"ontology_id", ontologyID,
			"error", err,
		)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "FEED_FAILED",
			Message: "Failed to list comment feed",
		})
		return
	}

	responses := make([]CommentResponse, 0, len(comments))
	for _, c := range comments {
		responses = append(responses, *toCommentResponse(&c, nil))
	}

	slog.Debug("Listed comment feed",
		"ontology_id", ontologyID,
		"total", total,
	)

	writeJSON(w, http.StatusOK, CommentFeedResponse{
		Comments: responses,
		Total:    total,
		Since:    since,
	})
}

// sseCommentStream handles GET /api/v1/ontologies/{ontology_id}/comments/stream
// It upgrades the connection to Server-Sent Events for live comment streaming.
func (h *commentHandlers) sseCommentStream(w http.ResponseWriter, r *http.Request) {
	ontologyID := r.PathValue("ontology_id")

	if h.bus == nil {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "SSE_UNAVAILABLE",
			Message: "Event bus is not available",
		})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "STREAMING_UNSUPPORTED",
			Message: "Streaming not supported by this connection",
		})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Send initial keepalive
	_, _ = fmt.Fprintf(w, ": connected to comment stream for ontology %s\n\n", ontologyID)
	flusher.Flush()

	broker := newSSEBroker(h.bus)
	eventCh := make(chan SSEEvent, 50)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go broker.ServeSSE(ctx, ontologyID, eventCh, errCh)

	slog.Info("SSE stream started", "ontology_id", ontologyID)

	for {
		select {
		case <-ctx.Done():
			slog.Info("SSE stream ended", "ontology_id", ontologyID)
			return
		case event, ok := <-eventCh:
			if !ok {
				return
			}
			_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, event.Data)
			if err != nil {
				slog.Warn("SSE write error", "ontology_id", ontologyID, "error", err)
				return
			}
			flusher.Flush()
		}
	}
}

// registerCommentRoutes registers all comment CRUD routes on the given mux.
func registerCommentRoutes(mux *http.ServeMux, handlers *commentHandlers) {
	// Comment CRUD
	mux.HandleFunc("POST /api/v1/ontologies/{ontology_id}/comments", handlers.createComment)
	mux.HandleFunc("GET /api/v1/comments/{id}", handlers.getComment)
	mux.HandleFunc("PUT /api/v1/comments/{id}", handlers.updateComment)
	mux.HandleFunc("DELETE /api/v1/comments/{id}", handlers.deleteComment)

	// Entity-scoped listing
	mux.HandleFunc("GET /api/v1/ontologies/{ontology_id}/comments", handlers.listCommentsByEntity)

	// Project-wide feed
	mux.HandleFunc("GET /api/v1/ontologies/{ontology_id}/comment-feed", handlers.listCommentFeed)

	// SSE live stream
	mux.HandleFunc("GET /api/v1/ontologies/{ontology_id}/comments/stream", handlers.sseCommentStream)
}

// toCommentResponse converts a Comment model to an API response, optionally embedding replies.
func toCommentResponse(c *Comment, replies []Comment) *CommentResponse {
	resp := &CommentResponse{
		ID:              c.ID,
		OntologyID:      c.OntologyID,
		EntityID:        c.EntityID,
		EntityType:      c.EntityType,
		AuthorID:        c.AuthorID,
		Body:            c.Body,
		ParentCommentID: c.ParentCommentID,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}

	if len(replies) > 0 {
		resp.Replies = make([]CommentResponse, 0, len(replies))
		for _, r := range replies {
			resp.Replies = append(resp.Replies, *toCommentResponse(&r, nil))
		}
	}

	return resp
}

// parseIntParam parses an integer query parameter with a default fallback.
func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}
