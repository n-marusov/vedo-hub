// Validates: REQ-FUN.INTEGRATION.collaboration
//
//go:build integration
// +build integration

// Integration tests for the commenting service with PostgreSQL-backed CommentStore.
//
// Requires a running PostgreSQL instance. Set COMMENTING_TEST_DATABASE_URL env var
// to specify the connection (defaults to localhost:5432/vedo_comments_test).
// If the database is unreachable, tests will skip gracefully.
//
// Run with:
//
//	cd src/services/commenting-service
//	go test -tags=integration -run TestIT_ -v .
//
// References: REQ-FUN.INTEGRATION.collaboration, US-team.comments.feed
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// ─── test helpers ──────────────────────────────────────────────────────────────

func getTestDBURL() string {
	if dsn := os.Getenv("COMMENTING_TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgres://postgres:password@localhost:5432/vedo_comments_test?sslmode=disable"
}

func setupITMux(t *testing.T) (*http.ServeMux, *CommentStore) {
	t.Helper()

	store, err := NewCommentStore(getTestDBURL())
	if err != nil {
		t.Skipf("PostgreSQL not available, skipping integration test: %v\nSet COMMENTING_TEST_DATABASE_URL if not using default.", err)
	}

	mux := http.NewServeMux()
	handlers := &commentHandlers{store: store, bus: newEventBus()}

	// Register cleanup to close the store when the test finishes
	t.Cleanup(func() { store.Close() })

	// Info + health endpoints
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	registerCommentRoutes(mux, handlers)

	// Clean previous test data
	_, _ = store.db.ExecContext(context.Background(),
		"DELETE FROM comments WHERE ontology_id LIKE 'it-test-%'")

	return mux, store
}

func itReq(mux *http.ServeMux, method, path string, body []byte) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	mux.ServeHTTP(w, req)
	return w
}

func waitForPG(t *testing.T, store *CommentStore) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		if err := store.db.PingContext(ctx); err == nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("PostgreSQL not reachable within timeout")
		case <-time.After(300 * time.Millisecond):
		}
	}
}

// ─── CRUD Integration Tests ───────────────────────────────────────────────────

func TestIT_CreateComment_StoresInPostgres(t *testing.T) {
	mux, store := setupITMux(t)
	waitForPG(t, store)

	ontologyID := "it-test-create"
	entityID := "class-it-1"
	body := fmt.Sprintf(
		`{"entity_id":%q,"entity_type":"class","body":"IT create test %s"}`,
		entityID, time.Now().Format(time.RFC3339),
	)

	w := itReq(mux, "POST", fmt.Sprintf("/api/v1/ontologies/%s/comments", ontologyID), []byte(body))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var resp CommentResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty comment ID")
	}
	if resp.EntityID != entityID {
		t.Errorf("expected entity_id %q, got %q", entityID, resp.EntityID)
	}

	// Verify directly in PostgreSQL
	stored, err := store.GetComment(context.Background(), resp.ID)
	if err != nil {
		t.Fatalf("GetComment failed: %v", err)
	}
	if stored == nil {
		t.Fatal("expected comment to exist in PostgreSQL, got nil")
	}
	if stored.Body != resp.Body {
		t.Errorf("body mismatch: expected %q, got %q", resp.Body, stored.Body)
	}
}

func TestIT_CreateCommentWithReply_AndFetchFeed(t *testing.T) {
	mux, store := setupITMux(t)
	waitForPG(t, store)

	ontologyID := "it-test-feed"
	entityID := "class-feed-it"

	// Step 1: Create top-level comment
	createBody := fmt.Sprintf(
		`{"entity_id":%q,"entity_type":"class","body":"Top-level comment for feed"}`,
		entityID,
	)
	w := itReq(mux, "POST", fmt.Sprintf("/api/v1/ontologies/%s/comments", ontologyID), []byte(createBody))
	if w.Code != http.StatusCreated {
		t.Fatalf("step 1: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var parent CommentResponse
	json.Unmarshal(w.Body.Bytes(), &parent)

	// Step 2: Create reply
	replyBody := fmt.Sprintf(
		`{"entity_id":%q,"entity_type":"class","body":"This is a reply","parent_id":%q}`,
		entityID, parent.ID,
	)
	w = itReq(mux, "POST", fmt.Sprintf("/api/v1/ontologies/%s/comments", ontologyID), []byte(replyBody))
	if w.Code != http.StatusCreated {
		t.Fatalf("step 2: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var reply CommentResponse
	json.Unmarshal(w.Body.Bytes(), &reply)
	if reply.ParentCommentID == nil || *reply.ParentCommentID != parent.ID {
		t.Errorf("expected parent_id %q, got %v", parent.ID, reply.ParentCommentID)
	}

	// Step 3: Entity-scoped listing
	listW := itReq(mux, "GET",
		fmt.Sprintf("/api/v1/ontologies/%s/comments?entity_id=%s", ontologyID, entityID),
		nil,
	)
	if listW.Code != http.StatusOK {
		t.Fatalf("step 3: expected 200, got %d: %s", listW.Code, listW.Body.String())
	}
	var listResp struct {
		Comments []CommentResponse `json:"comments"`
		Total    int               `json:"total"`
	}
	json.Unmarshal(listW.Body.Bytes(), &listResp)
	if listResp.Total < 1 {
		t.Error("expected at least 1 comment in list response")
	}

	// Step 4: Project-wide feed
	feedW := itReq(mux, "GET",
		fmt.Sprintf("/api/v1/ontologies/%s/comment-feed?since=2024-01-01T00:00:00Z", ontologyID),
		nil,
	)
	if feedW.Code != http.StatusOK {
		t.Fatalf("step 4: expected 200, got %d: %s", feedW.Code, feedW.Body.String())
	}
}

func TestIT_UpdateComment_PersistsChange(t *testing.T) {
	mux, store := setupITMux(t)
	waitForPG(t, store)

	ontologyID := "it-test-update"
	entityID := "class-update-it"

	// Create
	w := itReq(mux, "POST", fmt.Sprintf("/api/v1/ontologies/%s/comments", ontologyID),
		[]byte(fmt.Sprintf(`{"entity_id":%q,"entity_type":"class","body":"Original body"}`, entityID)))
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", w.Code)
	}
	var created CommentResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	// Update
	w = itReq(mux, "PUT", fmt.Sprintf("/api/v1/comments/%s", created.ID),
		[]byte(`{"body":"Updated body content"}`))
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var updated CommentResponse
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.Body != "Updated body content" {
		t.Errorf("expected 'Updated body content', got %q", updated.Body)
	}

	// Verify in PostgreSQL
	stored, _ := store.GetComment(context.Background(), created.ID)
	if stored == nil {
		t.Fatal("comment should exist after update")
	}
	if stored.Body != "Updated body content" {
		t.Errorf("PG body: expected %q, got %q", "Updated body content", stored.Body)
	}
}

func TestIT_DeleteComment_RemovesFromPostgres(t *testing.T) {
	mux, store := setupITMux(t)
	waitForPG(t, store)

	ontologyID := "it-test-delete"
	entityID := "class-delete-it"

	// Create
	w := itReq(mux, "POST", fmt.Sprintf("/api/v1/ontologies/%s/comments", ontologyID),
		[]byte(fmt.Sprintf(`{"entity_id":%q,"entity_type":"class","body":"To be deleted"}`, entityID)))
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", w.Code)
	}
	var created CommentResponse
	json.Unmarshal(w.Body.Bytes(), &created)

	// Delete
	w = itReq(mux, "DELETE", fmt.Sprintf("/api/v1/comments/%s", created.ID), nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify removed
	stored, _ := store.GetComment(context.Background(), created.ID)
	if stored != nil {
		t.Error("comment should be nil after deletion but was found in PostgreSQL")
	}
}

func TestIT_CommentFeed_MultipleComments(t *testing.T) {
	mux, store := setupITMux(t)
	waitForPG(t, store)

	ontologyID := "it-test-multi-feed"
	entityID := "class-multi-it"

	// Create several comments
	for i := 0; i < 3; i++ {
		w := itReq(mux, "POST", fmt.Sprintf("/api/v1/ontologies/%s/comments", ontologyID),
			[]byte(fmt.Sprintf(`{"entity_id":%q,"entity_type":"class","body":"Comment #%d"}`, entityID, i+1)))
		if w.Code != http.StatusCreated {
			t.Fatalf("create comment %d: expected 201, got %d", i+1, w.Code)
		}
	}

	// Fetch feed with limit
	feedW := itReq(mux, "GET",
		fmt.Sprintf("/api/v1/ontologies/%s/comment-feed?limit=5", ontologyID),
		nil,
	)
	if feedW.Code != http.StatusOK {
		t.Fatalf("feed: expected 200, got %d: %s", feedW.Code, feedW.Body.String())
	}
	var feedResp struct {
		Comments []CommentResponse `json:"comments"`
		Total    int               `json:"total"`
	}
	json.Unmarshal(feedW.Body.Bytes(), &feedResp)
	if feedResp.Total < 3 {
		t.Errorf("expected >=3 comments in feed, got %d", feedResp.Total)
	}
}
