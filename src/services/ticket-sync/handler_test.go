// @ctx: HTTP handler tests — contract test scenarios from TICKET-SYNC-001 test spec
// @hlv:artifact tests-ticket-sync verifies TICKET-SYNC-001

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *SyncEngine, *StubGitLabClient) {
	gin.SetMode(gin.TestMode)
	client := NewStubGitLabClient()
	engine := NewSyncEngine(client)
	handler := NewTicketSyncHandler(engine)

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(gin.Recovery())
	router.POST("ticket-sync", handler.SyncTicket)
	router.POST("webhook/gitlab", WebhookHMACMiddleware(), handler.WebhookGitLab)

	return router, engine, client
}

func execReq(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "test-trace-ticket-sync")
	router.ServeHTTP(w, req)
	return w
}

func execReqWithHeaders(router *gin.Engine, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	router.ServeHTTP(w, req)
	return w
}

func validTicketData() map[string]interface{} {
	return map[string]interface{}{
		"id":              "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"title":           "Test ticket for sync",
		"description":     "Testing synchronization to GitLab",
		"category":        "bug",
		"user_severity":   "medium",
		"system_priority": "p2",
		"source":          "manual",
		"channel":         "ui",
		"status":          "new",
	}
}

func validRequest(action string) string {
	data := validTicketData()
	dataJSON, _ := json.Marshal(data)
	return `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"action": "` + action + `",
		"ticket_data": ` + string(dataJSON) + `,
		"gitlab_project_id": "vedo-core/support"
	}`
}

// @hlv CT-TICKET-SYNC-001-001
func TestCT_CreateGitLabIssue(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := validRequest("create")
	w := execReq(router, "POST", "/ticket-sync", body)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp SyncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.SyncStatus != SyncStatusSynced {
		t.Errorf("expected sync_status=synced, got %s", resp.SyncStatus)
	}
	if resp.GitLabIssueID == nil || *resp.GitLabIssueID == "" {
		t.Errorf("expected gitlab_issue_id to be set")
	}
	if resp.GitLabIssueIID == nil || *resp.GitLabIssueIID == "" {
		t.Errorf("expected gitlab_issue_iid to be set")
	}
	if resp.GitLabIssueURL == nil || *resp.GitLabIssueURL == "" {
		t.Errorf("expected gitlab_issue_url to be set")
	}
	if !strings.Contains(*resp.GitLabIssueURL, "vedo-core/support/-/issues/") {
		t.Errorf("expected gitlab_issue_url to contain project path")
	}
	if resp.Error != nil {
		t.Errorf("expected no error, got %s", *resp.Error)
	}
	if resp.SyncID == "" {
		t.Errorf("expected sync_id to be set")
	}
	if resp.TicketID != "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8" {
		t.Errorf("expected ticket_id to match")
	}
}

// @hlv CT-TICKET-SYNC-001-002
func TestCT_UpdateGitLabIssue(t *testing.T) {
	router, engine, _ := setupTestRouter()

	createBody := validRequest("create")
	w1 := execReq(router, "POST", "/ticket-sync", createBody)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for create, got %d: %s", w1.Code, w1.Body.String())
	}
	var createResp SyncResponse
	json.Unmarshal(w1.Body.Bytes(), &createResp)

	record := engine.getRecord("a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8")
	if record == nil {
		t.Fatal("expected sync record to exist")
	}
	record.Status = SyncStatusSynced

	updateData := validTicketData()
	updateData["status"] = "resolved"
	updateDataJSON, _ := json.Marshal(updateData)
	updateBody := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"action": "update",
		"ticket_data": ` + string(updateDataJSON) + `,
		"gitlab_project_id": "vedo-core/support"
	}`

	w2 := execReq(router, "POST", "/ticket-sync", updateBody)
	if w2.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for update, got %d: %s", w2.Code, w2.Body.String())
	}

	var updateResp SyncResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &updateResp); err != nil {
		t.Fatalf("failed to parse update response: %v", err)
	}
	if updateResp.SyncStatus != SyncStatusSynced {
		t.Errorf("expected sync_status=synced, got %s", updateResp.SyncStatus)
	}
	if updateResp.GitLabIssueID == nil || *updateResp.GitLabIssueID != *createResp.GitLabIssueID {
		t.Errorf("expected gitlab_issue_id to be same as create, got %v vs %v",
			updateResp.GitLabIssueID, createResp.GitLabIssueID)
	}
	if updateResp.Error != nil {
		t.Errorf("expected no error, got %s", *updateResp.Error)
	}
}

// @hlv CT-TICKET-SYNC-001-003
func TestCT_InvalidTicketID(t *testing.T) {
	router, _, _ := setupTestRouter()
	data := validTicketData()
	dataJSON, _ := json.Marshal(data)
	body := `{
		"ticket_id": "not-a-valid-uuid",
		"action": "create",
		"ticket_data": ` + string(dataJSON) + `,
		"gitlab_project_id": "vedo-core/support"
	}`
	w := execReq(router, "POST", "/ticket-sync", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if errResp.Error != ErrInvalidRequest {
		t.Errorf("expected SYNC-INVALID-REQUEST, got %s", errResp.Error)
	}
	if errResp.Message == "" {
		t.Errorf("expected error message to be set")
	}
}

// @hlv CT-TICKET-SYNC-001-004
func TestCT_GitLabUnavailable(t *testing.T) {
	router, engine, client := setupTestRouter()

	client.SetSimulateFailure("unavailable")

	body := validRequest("create")
	w := execReq(router, "POST", "/ticket-sync", body)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if errResp.Error != ErrGitLabUnavailable {
		t.Errorf("expected SYNC-GITLAB-UNAVAILABLE, got %s", errResp.Error)
	}

	_ = engine
}

// @hlv CT-TICKET-SYNC-001-005
func TestCT_SyncConflict(t *testing.T) {
	router, engine, client := setupTestRouter()

	createBody := validRequest("create")
	w1 := execReq(router, "POST", "/ticket-sync", createBody)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for create, got %d: %s", w1.Code, w1.Body.String())
	}

	record := engine.getRecord("a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8")
	if record == nil {
		t.Fatal("expected sync record to exist")
	}
	// @hlv SYNC-CONFLICT — simulate externally closed issue
	record.Status = SyncStatusConflict

	updateData := validTicketData()
	updateData["status"] = "in_progress"
	updateDataJSON, _ := json.Marshal(updateData)
	updateBody := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"action": "update",
		"ticket_data": ` + string(updateDataJSON) + `,
		"gitlab_project_id": "vedo-core/support"
	}`

	w2 := execReq(router, "POST", "/ticket-sync", updateBody)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w2.Code, w2.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse conflict response: %v", err)
	}
	if errResp.Error != ErrConflict {
		t.Errorf("expected SYNC-CONFLICT, got %s", errResp.Error)
	}

	_ = client
}

// @hlv CT invalid body — missing required fields
func TestCT_MissingRequiredFields(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := `{"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8"}`
	w := execReq(router, "POST", "/ticket-sync", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// @hlv CT webhook — HMAC signature validation
func TestCT_WebhookHMACMissing(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := `{"event": "test"}`
	w := execReq(router, "POST", "/webhook/gitlab", body)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

// @hlv CT webhook — HMAC signature valid
func TestCT_WebhookHMACValid(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := `{"event": "push", "ref": "main"}`
	w := execReqWithHeaders(router, "POST", "/webhook/gitlab", body, map[string]string{
		"X-Gitlab-Token": computeHMAC(body),
	})
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}
}

// @hlv CT webhook — HMAC signature invalid
func TestCT_WebhookHMACInvalid(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := `{"event": "push"}`
	w := execReqWithHeaders(router, "POST", "/webhook/gitlab", body, map[string]string{
		"X-Gitlab-Token": "invalid-signature",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func computeHMAC(body string) string {
	mac := hmac.New(sha256.New, []byte(webhookSecretKey))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

// @hlv PBT-TICKET-SYNC-001-001 — Sync ID uniqueness property
func TestPBT_SyncIDUniqueness(t *testing.T) {
	router, _, client := setupTestRouter()

	ids := make(map[string]bool)
	actions := []string{"create", "update", "close", "reopen"}
	for i := 0; i < 50; i++ {
		ticketID := NewUUID()
		action := actions[i%len(actions)]
		data := validTicketData()
		data["id"] = ticketID
		dataJSON, _ := json.Marshal(data)
		body := `{
			"ticket_id": "` + ticketID + `",
			"action": "` + action + `",
			"ticket_data": ` + string(dataJSON) + `,
			"gitlab_project_id": "vedo-core/support"
		}`

		if action != "create" && i > 0 {
			client.ClearSimulateFailure()
		}

		w := execReq(router, "POST", "/ticket-sync", body)
		if w.Code == http.StatusAccepted || w.Code == http.StatusConflict || w.Code == http.StatusServiceUnavailable {
			var resp SyncResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil && resp.SyncID != "" {
				if ids[resp.SyncID] {
					t.Errorf("duplicate sync_id detected: %s", resp.SyncID)
				}
				ids[resp.SyncID] = true
			}
		}
	}
	if len(ids) < 10 {
		t.Errorf("expected at least 10 unique sync IDs, got %d", len(ids))
	}
}

// @hlv PBT-TICKET-SYNC-001-002 — Sync status validity property
func TestPBT_SyncStatusValidity(t *testing.T) {
	router, _, client := setupTestRouter()

	validStatuses := map[SyncStatus]bool{
		SyncStatusPending:    true,
		SyncStatusInProgress: true,
		SyncStatusSynced:     true,
		SyncStatusFailed:     true,
		SyncStatusConflict:   true,
	}

	expectedErrors := map[string]string{
		"unavailable":  ErrGitLabUnavailable,
		"rate_limited": ErrRateLimited,
		"conflict":     ErrConflict,
	}

	// Good path — verify synced status
	client.ClearSimulateFailure()
	ticketID0 := NewUUID()
	data0 := validTicketData()
	data0["id"] = ticketID0
	dataJSON0, _ := json.Marshal(data0)
	body0 := `{
		"ticket_id": "` + ticketID0 + `",
		"action": "create",
		"ticket_data": ` + string(dataJSON0) + `,
		"gitlab_project_id": "vedo-core/support"
	}`
	w0 := execReq(router, "POST", "/ticket-sync", body0)
	var resp0 SyncResponse
	if err := json.Unmarshal(w0.Body.Bytes(), &resp0); err == nil {
		if !validStatuses[resp0.SyncStatus] {
			t.Errorf("invalid sync_status: %s", resp0.SyncStatus)
		}
		if resp0.SyncStatus != SyncStatusSynced {
			t.Errorf("expected synced for good path, got %s", resp0.SyncStatus)
		}
	}

	for mode, expectedErr := range expectedErrors {
		client.ClearSimulateFailure()
		client.SetSimulateFailure(mode)

		ticketID := NewUUID()
		data := validTicketData()
		data["id"] = ticketID
		dataJSON, _ := json.Marshal(data)
		body := `{
			"ticket_id": "` + ticketID + `",
			"action": "create",
			"ticket_data": ` + string(dataJSON) + `,
			"gitlab_project_id": "vedo-core/support"
		}`
		w := execReq(router, "POST", "/ticket-sync", body)

		var errResp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err == nil && errResp.Error != "" {
			if errResp.Error != expectedErr {
				t.Errorf("mode %s: expected error %s, got %s", mode, expectedErr, errResp.Error)
			}
		} else {
			var resp SyncResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
				if !validStatuses[resp.SyncStatus] {
					t.Errorf("mode %s: invalid sync_status: %s", mode, resp.SyncStatus)
				}
			}
		}
	}
}

// @hlv PBT-TICKET-SYNC-001-003 — GitLab issue ID consistency property
func TestPBT_GitLabIssueIDConsistency(t *testing.T) {
	router, engine, _ := setupTestRouter()

	ticketID := "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8"

	createBody := validRequest("create")
	w1 := execReq(router, "POST", "/ticket-sync", createBody)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for create, got %d: %s", w1.Code, w1.Body.String())
	}

	var createResp SyncResponse
	json.Unmarshal(w1.Body.Bytes(), &createResp)
	if createResp.GitLabIssueID == nil {
		t.Fatal("expected gitlab_issue_id from create")
	}

	firstIssueID := *createResp.GitLabIssueID
	firstIssueIID := *createResp.GitLabIssueIID

	record := engine.getRecord(ticketID)
	if record != nil {
		record.Status = SyncStatusSynced
	}

	for i := 0; i < 5; i++ {
		data := validTicketData()
		data["status"] = "in_progress"
		dataJSON, _ := json.Marshal(data)
		updateBody := `{
			"ticket_id": "` + ticketID + `",
			"action": "update",
			"ticket_data": ` + string(dataJSON) + `,
			"gitlab_project_id": "vedo-core/support"
		}`

		w := execReq(router, "POST", "/ticket-sync", updateBody)
		if w.Code != http.StatusAccepted {
			t.Fatalf("expected 202 for update iteration %d, got %d", i, w.Code)
		}

		var resp SyncResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.GitLabIssueID == nil || *resp.GitLabIssueID != firstIssueID {
			t.Errorf("iteration %d: expected gitlab_issue_id=%s, got %v", i, firstIssueID, resp.GitLabIssueID)
		}
		if resp.GitLabIssueIID == nil || *resp.GitLabIssueIID != firstIssueIID {
			t.Errorf("iteration %d: expected gitlab_issue_iid=%s, got %v", i, firstIssueIID, resp.GitLabIssueIID)
		}
	}
}
