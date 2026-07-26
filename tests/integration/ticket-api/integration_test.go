// @ctx: integration tests for TICKET-CORE-001 — end-to-end lifecycle
// @hlv:artifact tests-ticket-api verifies TICKET-CORE-001

package ticketapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vedo-core/src/services/ticket-api/ticketapi"

	"github.com/gin-gonic/gin"
)

func setupApp() (*gin.Engine, *ticketapi.TicketStore, *ticketapi.AuditStore) {
	gin.SetMode(gin.TestMode)
	store := ticketapi.NewTicketStore()
	audit := ticketapi.NewAuditStore()
	handler := ticketapi.NewTicketHandler(store, audit, "test", "1.5.0")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("trace_id", "it-trace-001")
		c.Set("user_id", "test-user-uuid")
		c.Set("role", "support")
		c.Next()
	})

	tickets := router.Group("/api/v1/tickets")
	{
		tickets.POST("", handler.CreateTicket)
		tickets.GET("", handler.ListTickets)
		tickets.GET("/:id", handler.GetTicket)
		tickets.PATCH("/:id", handler.UpdateTicket)
		tickets.POST("/:id/comments", handler.AddComment)
		tickets.GET("/:id/audit", handler.GetAuditLog)
	}

	return router, store, audit
}

func itReq(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	return w
}

// @hlv IT-TICKET-CORE-001-001
func TestIT_EndToEndLifecycle(t *testing.T) {
	router, _, _ := setupApp()

	createBody := `{
		"title": "Ontology service down",
		"description": "Service returning 500 errors",
		"category": "bug",
		"user_severity": "high",
		"source": "manual",
		"channel": "ui"
	}`
	w := itReq(router, "POST", "/api/v1/tickets", createBody)
	if w.Code != http.StatusCreated {
		t.Fatalf("step 1: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created ticketapi.Ticket
	json.Unmarshal(w.Body.Bytes(), &created)
	ticketID := created.ID

	updateBody := `{"priority": "p0", "comment": "Critical outage escalation"}`
	w = itReq(router, "PATCH", "/api/v1/tickets/"+ticketID, updateBody)
	if w.Code != http.StatusOK {
		t.Fatalf("step 2: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var updated ticketapi.Ticket
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.SystemPriority != ticketapi.TicketPriorityP0 {
		t.Errorf("expected p0, got %s", updated.SystemPriority)
	}

	w = itReq(router, "GET", "/api/v1/tickets/"+ticketID, "")
	if w.Code != http.StatusOK {
		t.Fatalf("step 3: expected 200, got %d", w.Code)
	}

	closeBody := `{"status": "closed"}`
	w = itReq(router, "PATCH", "/api/v1/tickets/"+ticketID, closeBody)
	if w.Code != http.StatusOK {
		t.Fatalf("step 4: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var closed ticketapi.Ticket
	json.Unmarshal(w.Body.Bytes(), &closed)
	if closed.Status != ticketapi.TicketStatusClosed {
		t.Errorf("expected closed, got %s", closed.Status)
	}

	w = itReq(router, "GET", "/api/v1/tickets/"+ticketID+"/audit", "")
	if w.Code != http.StatusOK {
		t.Fatalf("step 5: expected 200, got %d", w.Code)
	}
	var auditResp struct {
		Entries []ticketapi.AuditEntry `json:"entries"`
		Total   int                    `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &auditResp)
	if auditResp.Total < 2 {
		t.Errorf("expected >=2 audit entries, got %d", auditResp.Total)
	}
}
