// @ctx: HTTP handler tests — contract test scenarios from TICKET-CORE-001 test spec
// @hlv:artifact tests-ticket-api verifies TICKET-CORE-001

package ticketapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *TicketStore, *AuditStore) {
	gin.SetMode(gin.TestMode)
	store := NewTicketStore()
	audit := NewAuditStore()
	handler := NewTicketHandler(store, audit, "test", "1.5.0")

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(gin.Recovery())

	tickets := router.Group("/api/v1/tickets")
	{
		tickets.POST("", handler.CreateTicket)
		tickets.GET("", handler.ListTickets)
		tickets.GET("/:id", handler.GetTicket)
		tickets.PATCH("/:id", handler.UpdateTicket)
		tickets.POST("/:id/comments", handler.AddComment)
		tickets.GET("/:id/audit", handler.GetAuditLog)
	}
	router.GET("/health", handler.HealthCheck)

	return router, store, audit
}

func execReq(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "test-trace-001")
	req.Header.Set("X-User-Id", "test-user-uuid")
	req.Header.Set("X-User-Role", "support")
	router.ServeHTTP(w, req)
	return w
}

// @hlv CT-TICKET-CORE-001-001
func TestCT_CreateManualTicket(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := `{
		"title": "Test ticket creation",
		"description": "Creating a test ticket to verify the API works",
		"category": "bug",
		"user_severity": "medium",
		"source": "manual",
		"channel": "ui"
	}`
	w := execReq(router, "POST", "/api/v1/tickets", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp Ticket
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Source != TicketSourceManual {
		t.Errorf("expected source=manual, got %s", resp.Source)
	}
	if resp.Channel != TicketChannelUI {
		t.Errorf("expected channel=ui, got %s", resp.Channel)
	}
	if resp.Status != TicketStatusNew {
		t.Errorf("expected status=new, got %s", resp.Status)
	}
	if resp.Title != "Test ticket creation" {
		t.Errorf("expected title mismatch")
	}
	if resp.SystemPriority != TicketPriorityP2 {
		t.Errorf("expected priority=p2, got %s", resp.SystemPriority)
	}
	if len(resp.Labels) != 0 {
		t.Errorf("expected 0 labels for manual ticket, got %d", len(resp.Labels))
	}
}

// @hlv CT-TICKET-CORE-001-002
func TestCT_CreateTelemetryTicketWithAttachment(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := `{
		"title": "High latency detected in SPARQL endpoint",
		"description": "SPARQL query latency exceeded threshold for 5 consecutive minutes",
		"category": "performance",
		"user_severity": "high",
		"attachments": [{"filename": "latency_metrics.csv", "size": 2048, "url": "https://storage.example.com/telemetry/latency_metrics.csv"}],
		"source": "telemetry",
		"channel": "telemetry"
	}`
	w := execReq(router, "POST", "/api/v1/tickets", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp Ticket
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Source != TicketSourceTelemetry {
		t.Errorf("expected source=telemetry, got %s", resp.Source)
	}
	if resp.Channel != TicketChannelTelemetry {
		t.Errorf("expected channel=telemetry, got %s", resp.Channel)
	}
	if resp.SystemPriority != TicketPriorityP1 {
		t.Errorf("expected priority=p1, got %s", resp.SystemPriority)
	}
	if len(resp.Attachments) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(resp.Attachments))
	}
	hasSourceLabel := false
	for _, l := range resp.Labels {
		if l == "source=telemetry" {
			hasSourceLabel = true
			break
		}
	}
	if !hasSourceLabel {
		t.Error("expected telemetry label 'source=telemetry'")
	}
}

// @hlv CT-TICKET-CORE-001-003
func TestCT_MissingRequiredField(t *testing.T) {
	router, _, _ := setupTestRouter()
	body := `{
		"description": "Ticket without title",
		"category": "bug",
		"user_severity": "low",
		"source": "manual",
		"channel": "ui"
	}`
	w := execReq(router, "POST", "/api/v1/tickets", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error: %v", err)
	}
	if errResp.Error != ErrInvalidRequest {
		t.Errorf("expected TICKET-INVALID-REQUEST, got %s", errResp.Error)
	}
}

// @hlv CT-TICKET-CORE-001-004
func TestCT_TooManyAttachments(t *testing.T) {
	router, _, _ := setupTestRouter()
	var atts []string
	for i := 0; i < 11; i++ {
		atts = append(atts, `{"filename":"file`+string(rune('0'+i))+`.txt","size":100,"url":"https://example.com/file`+string(rune('0'+i))+`.txt"}`)
	}
	body := `{
		"title": "Ticket with excessive attachments",
		"description": "Testing attachment limit",
		"category": "question",
		"user_severity": "low",
		"attachments": [` + strings.Join(atts, ",") + `],
		"source": "manual",
		"channel": "ui"
	}`
	w := execReq(router, "POST", "/api/v1/tickets", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error: %v", err)
	}
	if errResp["error"] != ErrTooManyAttachments {
		t.Errorf("expected TICKET-TOO-MANY-ATTACHMENTS, got %s", errResp["error"])
	}
}

// @hlv CT-TICKET-CORE-001-005
func TestCT_UpdateTicketPriorityViaCLI(t *testing.T) {
	router, store, _ := setupTestRouter()

	ticket, _ := store.Create(CreateTicketRequest{
		Title: "Existing", Description: "Existing desc",
		Category: TicketCategoryBug, UserSeverity: TicketSeverityMedium,
		Source: TicketSourceManual, Channel: TicketChannelUI,
	}, TicketMetadata{UserID: "support-uuid", Environment: "test"})

	body := `{"priority": "p0"}`
	w := execReq(router, "PATCH", "/api/v1/tickets/"+ticket.ID, body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp Ticket
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.SystemPriority != TicketPriorityP0 {
		t.Errorf("expected priority=p0, got %s", resp.SystemPriority)
	}
	if resp.ID != ticket.ID {
		t.Errorf("expected same ticket ID")
	}
}

func TestCT_NotFound(t *testing.T) {
	// @hlv TICKET-NOT-FOUND
	router, _, _ := setupTestRouter()
	w := execReq(router, "GET", "/api/v1/tickets/nonexistent-uuid", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCT_AddComment(t *testing.T) {
	router, store, _ := setupTestRouter()
	ticket, _ := store.Create(CreateTicketRequest{
		Title: "Test", Description: "Desc",
		Category: TicketCategoryBug, UserSeverity: TicketSeverityLow,
		Source: TicketSourceManual, Channel: TicketChannelUI,
	}, TicketMetadata{})

	body := `{"text": "This is a test comment"}`
	w := execReq(router, "POST", "/api/v1/tickets/"+ticket.ID+"/comments", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCT_AuditTrail(t *testing.T) {
	// @hlv audit_trail
	router, store, audit := setupTestRouter()
	ticket, _ := store.Create(CreateTicketRequest{
		Title: "Test", Description: "Desc",
		Category: TicketCategoryBug, UserSeverity: TicketSeverityMedium,
		Source: TicketSourceManual, Channel: TicketChannelUI,
	}, TicketMetadata{UserID: "user-uuid", Environment: "test"})

	audit.Record("user-uuid", "support", "create", ticket.ID, "trace-001", "success")
	audit.Record("user-uuid", "support", "update", ticket.ID, "trace-001", "success")

	w := execReq(router, "GET", "/api/v1/tickets/"+ticket.ID+"/audit", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Entries []AuditEntry `json:"entries"`
		Total   int          `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse audit response: %v", err)
	}
	if resp.Total < 2 {
		t.Errorf("expected at least 2 audit entries, got %d", resp.Total)
	}
}

func TestHealthEndpoint(t *testing.T) {
	router, _, _ := setupTestRouter()
	w := execReq(router, "GET", "/health", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
