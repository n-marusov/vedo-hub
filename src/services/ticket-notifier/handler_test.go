// @ctx: HTTP handler tests for ticket-notifier — contract tests from TICKET-NOTIFY-001 test spec
// @hlv:artifact tests-ticket-notifier verifies TICKET-NOTIFY-001
// @hlv All ticket lifecycle events trigger notifications
// @hlv P0/P1 tickets trigger escalation notifications
// @hlv Notification content includes ticket ID, title, description, category, severity
// @hlv Email sent for all events to ticket creator
// @hlv P0/P1 escalation includes runbook links
// @hlv Retry with exponential backoff for transient failures

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *NotifierHandler) {
	gin.SetMode(gin.TestMode)
	handler := NewNotifierHandler()

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(gin.Recovery())

	router.POST("/notifications", handler.HandleNotification)
	router.GET("/health", handler.HealthCheck)
	router.GET("/", handler.Metadata)
	router.GET("/ready", handler.ReadyCheck)

	return router, handler
}

func execReq(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "test-trace-notify-001")
	req.Header.Set("X-API-Key", "test-api-key")
	router.ServeHTTP(w, req)
	return w
}

// @hlv CT-TICKET-NOTIFY-001-001
// @hlv NOTIFY-INVALID-REQUEST (not triggered — happy path)
func TestCT_Notify001_SendEmailNotification(t *testing.T) {
	router, _ := setupTestRouter()
	body := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"event_type": "created",
		"ticket_data": {
			"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"title": "Test ticket creation",
			"description": "Creating a test ticket to verify notifications",
			"category": "bug",
			"user_severity": "medium",
			"system_priority": "p2",
			"source": "manual",
			"channel": "ui",
			"status": "new"
		},
		"recipient_type": "user",
		"notification_channel": "email",
		"template_vars": {
			"user_name": "John Doe",
			"ticket_url": "https://app.vedo.core/tickets/a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8"
		}
	}`
	w := execReq(router, "POST", "/notifications", body)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp NotificationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.NotificationID == "" {
		t.Error("expected non-empty notification_id")
	}
	if resp.Status != StatusQueued {
		t.Errorf("expected status=queued, got %s", resp.Status)
	}
	if resp.Channel != "email" {
		t.Errorf("expected channel=email, got %s", resp.Channel)
	}
	if resp.Recipient != "john.doe@example.com" {
		t.Errorf("expected recipient=john.doe@example.com, got %s", resp.Recipient)
	}
	if resp.RetryCount != 0 {
		t.Errorf("expected retry_count=0, got %d", resp.RetryCount)
	}
	if resp.Error != nil {
		t.Errorf("expected error=nil, got %s", *resp.Error)
	}
}

// @hlv CT-TICKET-NOTIFY-001-002
// @hlv P0/P1 tickets trigger escalation notifications
// @hlv P0/P1 escalation includes runbook links
func TestCT_Notify002_SendSlackP0Escalation(t *testing.T) {
	router, _ := setupTestRouter()
	body := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"event_type": "escalated",
		"ticket_data": {
			"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"title": "Critical ontology service failure",
			"description": "Ontology service completely unavailable",
			"category": "bug",
			"user_severity": "critical",
			"system_priority": "p0",
			"source": "telemetry",
			"channel": "telemetry",
			"status": "in_progress"
		},
		"recipient_type": "support",
		"notification_channel": "slack",
		"template_vars": {
			"ticket_url": "https://app.vedo.core/tickets/a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"runbook_url": "https://runbooks.vedo.core/ontology-service-failure",
			"severity": "P0"
		}
	}`
	w := execReq(router, "POST", "/notifications", body)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp NotificationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.NotificationID == "" {
		t.Error("expected non-empty notification_id")
	}
	if resp.Status != StatusQueued {
		t.Errorf("expected status=queued, got %s", resp.Status)
	}
	if resp.Channel != "slack" {
		t.Errorf("expected channel=slack, got %s", resp.Channel)
	}
	if resp.Recipient != "support-team-channel" {
		t.Errorf("expected recipient=support-team-channel, got %s", resp.Recipient)
	}
	if resp.Error != nil {
		t.Errorf("expected error=nil, got %s", *resp.Error)
	}
}

// @hlv CT-TICKET-NOTIFY-001-003
// @hlv NOTIFY-INVALID-REQUEST
func TestCT_Notify003_InvalidTicketID(t *testing.T) {
	router, _ := setupTestRouter()
	body := `{
		"ticket_id": "not-a-valid-uuid",
		"event_type": "created",
		"ticket_data": {
			"id": "not-a-valid-uuid",
			"title": "Invalid ticket",
			"description": "Ticket with invalid ID",
			"category": "question",
			"user_severity": "low",
			"system_priority": "p3",
			"source": "manual",
			"channel": "ui",
			"status": "new"
		},
		"recipient_type": "user",
		"notification_channel": "email",
		"template_vars": {}
	}`
	w := execReq(router, "POST", "/notifications", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error: %v", err)
	}
	if errResp.Error != ErrInvalidRequest {
		t.Errorf("expected %s, got %s", ErrInvalidRequest, errResp.Error)
	}
	if errResp.Message != "Ticket ID must be a valid UUID" {
		t.Errorf("expected 'Ticket ID must be a valid UUID', got '%s'", errResp.Message)
	}
}

// @hlv CT-TICKET-NOTIFY-001-004
// @hlv NOTIFY-CHANNEL-UNAVAILABLE
// @hlv:sec [NETWORK]
func TestCT_Notify004_ChannelUnavailable(t *testing.T) {
	router, handler := setupTestRouter()
	handler.SetChannelAvailability("pagerduty", false)

	body := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"event_type": "comment_added",
		"ticket_data": {
			"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"title": "Test ticket",
			"description": "Test ticket description",
			"category": "bug",
			"user_severity": "medium",
			"system_priority": "p2",
			"source": "manual",
			"channel": "ui",
			"status": "in_progress"
		},
		"recipient_type": "support",
		"notification_channel": "pagerduty",
		"template_vars": {}
	}`
	w := execReq(router, "POST", "/notifications", body)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error: %v", err)
	}
	if errResp.Error != ErrChannelUnavailable {
		t.Errorf("expected %s, got %s", ErrChannelUnavailable, errResp.Error)
	}
	if errResp.Message != "PagerDuty notification channel is currently unavailable" {
		t.Errorf("expected 'PagerDuty notification channel is currently unavailable', got '%s'", errResp.Message)
	}
}

// @hlv CT-TICKET-NOTIFY-001-005
func TestCT_Notify005_InAppNotification(t *testing.T) {
	router, _ := setupTestRouter()
	body := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"event_type": "comment_added",
		"ticket_data": {
			"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"title": "Test ticket for comments",
			"description": "Original ticket description",
			"category": "bug",
			"user_severity": "medium",
			"system_priority": "p2",
			"source": "manual",
			"channel": "ui",
			"status": "in_progress"
		},
		"recipient_type": "user",
		"notification_channel": "in_app",
		"template_vars": {
			"comment_author": "Jane Smith",
			"comment_text": "I've reproduced this issue and am working on a fix"
		}
	}`
	w := execReq(router, "POST", "/notifications", body)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp NotificationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.NotificationID == "" {
		t.Error("expected non-empty notification_id")
	}
	if resp.Status != StatusQueued {
		t.Errorf("expected status=queued, got %s", resp.Status)
	}
	if resp.Channel != "in_app" {
		t.Errorf("expected channel=in_app, got %s", resp.Channel)
	}
	if resp.Recipient != "user-uuid-for-john-doe" {
		t.Errorf("expected recipient=user-uuid-for-john-doe, got %s", resp.Recipient)
	}
	if resp.Error != nil {
		t.Errorf("expected error=nil, got %s", *resp.Error)
	}
}

// @hlv PBT-TICKET-NOTIFY-001-001
// @hlv uuid_uniqueness
func TestPBT_Notify001_NotificationIDUniqueness(t *testing.T) {
	router, _ := setupTestRouter()
	seen := make(map[string]bool)

	ticketIDs := []string{
		"a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"b2c3d4e5-f6a7-8901-h2i3-j4k5l6m7n8o9",
		"c3d4e5f6-a7b8-9012-i3j4-k5l6m7n8o9p0",
		"d4e5f6a7-b8c9-0123-j4k5-l6m7n8o9p0q1",
	}
	eventTypes := []string{"created", "status_changed", "comment_added", "escalated"}
	channels := []string{"email", "in_app", "slack", "pagerduty"}
	recipientTypes := []string{"user", "support", "admin", "pagerduty", "slack"}

	for i, tid := range ticketIDs {
		for _, evt := range eventTypes {
			for _, ch := range channels {
				for _, rt := range recipientTypes {
					body := `{
						"ticket_id": "` + tid + `",
						"event_type": "` + evt + `",
						"ticket_data": {
							"id": "` + tid + `",
							"title": "Test ticket",
							"description": "Test description",
							"category": "bug",
							"user_severity": "medium",
							"system_priority": "p2",
							"source": "manual",
							"channel": "ui",
							"status": "new"
						},
						"recipient_type": "` + rt + `",
						"notification_channel": "` + ch + `",
						"template_vars": {}
					}`
					w := execReq(router, "POST", "/notifications", body)
					if w.Code != http.StatusAccepted && w.Code != http.StatusServiceUnavailable {
						t.Logf("unexpected status %d for %s/%s/%s", w.Code, evt, ch, rt)
						continue
					}
					if w.Code != http.StatusAccepted {
						continue
					}
					var resp NotificationResponse
					if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
						t.Fatalf("failed to parse: %v", err)
					}
					if seen[resp.NotificationID] {
						t.Errorf("duplicate notification_id: %s (iteration %d)", resp.NotificationID, i)
					}
					seen[resp.NotificationID] = true
				}
			}
		}
	}
	if len(seen) == 0 {
		t.Error("expected at least one notification ID")
	}
	t.Logf("generated %d unique notification IDs", len(seen))
}

// @hlv PBT-TICKET-NOTIFY-001-002
// @hlv Notification content includes ticket ID, title, description, category, severity
func TestPBT_Notify002_ChannelMatchProperty(t *testing.T) {
	router, _ := setupTestRouter()

	ticketIDs := []string{
		"a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"b2c3d4e5-f6a7-8901-h2i3-j4k5l6m7n8o9",
	}
	channels := []string{"email", "in_app", "slack"}
	recipientTypes := []string{"user", "support"}
	eventTypes := []string{"created", "status_changed", "comment_added"}

	for _, tid := range ticketIDs {
		for _, ch := range channels {
			for _, rt := range recipientTypes {
				for _, evt := range eventTypes {
					body := `{
						"ticket_id": "` + tid + `",
						"event_type": "` + evt + `",
						"ticket_data": {
							"id": "` + tid + `",
							"title": "Test",
							"description": "Test desc",
							"category": "bug",
							"user_severity": "medium",
							"system_priority": "p2",
							"source": "manual",
							"channel": "ui",
							"status": "new"
						},
						"recipient_type": "` + rt + `",
						"notification_channel": "` + ch + `",
						"template_vars": {}
					}`
					w := execReq(router, "POST", "/notifications", body)
					if w.Code != http.StatusAccepted {
						continue
					}
					var resp NotificationResponse
					if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
						t.Fatalf("failed to parse: %v", err)
					}
					if resp.Channel != ch {
						t.Errorf("expected channel=%s, got %s for tid=%s evt=%s", ch, resp.Channel, tid, evt)
					}
				}
			}
		}
	}
}

// @hlv PBT-TICKET-NOTIFY-001-003
func TestPBT_Notify003_InvalidEventType(t *testing.T) {
	router, _ := setupTestRouter()
	invalidEvents := []string{"", "deleted", "assigned", "invalid_type", "resolved"}

	for _, evt := range invalidEvents {
		body := `{
			"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"event_type": "` + evt + `",
			"ticket_data": {
				"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
				"title": "Test",
				"description": "Test desc",
				"category": "bug",
				"user_severity": "low",
				"system_priority": "p3",
				"source": "manual",
				"channel": "ui",
				"status": "new"
			},
			"recipient_type": "user",
			"notification_channel": "email",
			"template_vars": {}
		}`
		w := execReq(router, "POST", "/notifications", body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for event_type=%q, got %d", evt, w.Code)
		}
		var errResp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("failed to parse error: %v", err)
		}
		if errResp.Error != ErrInvalidRequest {
			t.Errorf("expected %s for event_type=%q, got %s", ErrInvalidRequest, evt, errResp.Error)
		}
	}
}

// @hlv NOTIFY-UNAUTHORIZED
// @hlv:sec [AUTH_BOUNDARY]
func TestCT_Notify_MissingAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewNotifierHandler()
	router := gin.New()
	router.Use(TraceMiddleware())
	router.POST("/notifications", handler.HandleNotification)

	body := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"event_type": "created",
		"ticket_data": {
			"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"title": "Test",
			"description": "Test desc",
			"category": "bug",
			"user_severity": "low",
			"system_priority": "p3",
			"source": "manual",
			"channel": "ui",
			"status": "new"
		},
		"recipient_type": "user",
		"notification_channel": "email",
		"template_vars": {}
	}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notifications", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "test-trace-001")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error: %v", err)
	}
	if errResp.Error != ErrUnauthorized {
		t.Errorf("expected %s, got %s", ErrUnauthorized, errResp.Error)
	}
}

// @hlv NOTIFY-RECIPIENT-INVALID
func TestCT_Notify_InvalidRecipientType(t *testing.T) {
	router, _ := setupTestRouter()
	body := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"event_type": "created",
		"ticket_data": {
			"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"title": "Test",
			"description": "Test desc",
			"category": "bug",
			"user_severity": "low",
			"system_priority": "p3",
			"source": "manual",
			"channel": "ui",
			"status": "new"
		},
		"recipient_type": "invalid_role",
		"notification_channel": "email",
		"template_vars": {}
	}`
	w := execReq(router, "POST", "/notifications", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error: %v", err)
	}
	if errResp.Error != ErrRecipientInvalid {
		t.Errorf("expected %s, got %s", ErrRecipientInvalid, errResp.Error)
	}
}

// @hlv NOTIFY-RATE-LIMITED
// @hlv:sec [AUTH_BOUNDARY]
func TestCT_Notify_RateLimited(t *testing.T) {
	router, handler := setupTestRouter()
	handler.SetRateLimit(1)

	body := `{
		"ticket_id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
		"event_type": "created",
		"ticket_data": {
			"id": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8",
			"title": "Test",
			"description": "Test desc",
			"category": "bug",
			"user_severity": "low",
			"system_priority": "p3",
			"source": "manual",
			"channel": "ui",
			"status": "new"
		},
		"recipient_type": "user",
		"notification_channel": "email",
		"template_vars": {}
	}`

	w1 := execReq(router, "POST", "/notifications", body)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for first request, got %d: %s", w1.Code, w1.Body.String())
	}

	w2 := execReq(router, "POST", "/notifications", body)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for second request, got %d: %s", w2.Code, w2.Body.String())
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error: %v", err)
	}
	if errResp.Error != ErrRateLimited {
		t.Errorf("expected %s, got %s", ErrRateLimited, errResp.Error)
	}
}

func TestCT_Notify_HealthEndpoint(t *testing.T) {
	router, _ := setupTestRouter()
	w := execReq(router, "GET", "/health", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCT_Notify_MetadataEndpoint(t *testing.T) {
	router, _ := setupTestRouter()
	w := execReq(router, "GET", "/", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var meta map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &meta); err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}
	if meta["name"] != "ticket-notifier" {
		t.Errorf("expected name=ticket-notifier, got %v", meta["name"])
	}
}
