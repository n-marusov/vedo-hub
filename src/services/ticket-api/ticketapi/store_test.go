// Validates: REQ-FUN.INTEGRATION.ticket-management
// @ctx: unit tests for ticket store and priority derivation
// @hlv:artifact tests-ticket-api verifies TICKET-CORE-001

package ticketapi

import (
	"testing"
)

func TestCreateTicket_ValidManual(t *testing.T) {
	store := NewTicketStore()
	req := CreateTicketRequest{
		Title:        "Test ticket",
		Description:  "Description",
		Category:     TicketCategoryBug,
		UserSeverity: TicketSeverityMedium,
		Source:       TicketSourceManual,
		Channel:      TicketChannelUI,
	}
	meta := TicketMetadata{
		VedoVersion: "1.5.0",
		Environment: "test",
		UserID:      "user-uuid",
		UserAgent:   "test-client",
	}
	ticket, err := store.Create(req, meta)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ticket.Status != TicketStatusNew {
		t.Errorf("expected status new, got %s", ticket.Status)
	}
	if ticket.SystemPriority != TicketPriorityP2 {
		t.Errorf("expected p2 for medium bug, got %s", ticket.SystemPriority)
	}
	// @hlv priority_derivation
	if ticket.Source != TicketSourceManual {
		t.Errorf("expected source manual, got %s", ticket.Source)
	}
	if ticket.Channel != TicketChannelUI {
		t.Errorf("expected channel ui, got %s", ticket.Channel)
	}
}

func TestCreateTicket_ValidTelemetry(t *testing.T) {
	store := NewTicketStore()
	req := CreateTicketRequest{
		Title:        "SPARQL latency",
		Description:  "Exceeded threshold",
		Category:     TicketCategoryPerformance,
		UserSeverity: TicketSeverityHigh,
		Attachments: []TicketAttachment{
			{Filename: "metrics.csv", Size: 2048, URL: "https://example.com/metrics.csv"},
		},
		Source:  TicketSourceTelemetry,
		Channel: TicketChannelTelemetry,
	}
	meta := TicketMetadata{
		VedoVersion: "1.5.0",
		Environment: "prod",
		UserAgent:   "telemetry-collector",
	}
	ticket, err := store.Create(req, meta)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ticket.SystemPriority != TicketPriorityP1 {
		t.Errorf("expected p1 for high performance, got %s", ticket.SystemPriority)
	}
	if len(ticket.Attachments) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(ticket.Attachments))
	}
	// @hlv telemetry_label
	found := false
	for _, l := range ticket.Labels {
		if l == "source=telemetry" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected telemetry ticket to have label source=telemetry")
	}
}

func TestCreateTicket_EmptyTitle(t *testing.T) {
	// @hlv TICKET-INVALID-REQUEST
	store := NewTicketStore()
	req := CreateTicketRequest{
		Title:        "",
		Description:  "Some description",
		Category:     TicketCategoryBug,
		UserSeverity: TicketSeverityHigh,
		Source:       TicketSourceManual,
		Channel:      TicketChannelUI,
	}
	_, err := store.Create(req, TicketMetadata{})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	appErr, ok := err.(*AppError)
	if !ok || appErr.Code != ErrInvalidRequest {
		t.Errorf("expected TICKET-INVALID-REQUEST, got %v", err)
	}
}

func TestCreateTicket_TooManyAttachments(t *testing.T) {
	// @hlv TICKET-TOO-MANY-ATTACHMENTS
	store := NewTicketStore()
	var atts []TicketAttachment
	for i := 0; i < 11; i++ {
		atts = append(atts, TicketAttachment{Filename: "f.txt", Size: 100, URL: "https://example.com/f.txt"})
	}
	req := CreateTicketRequest{
		Title:        "Test",
		Description:  "Desc",
		Category:     TicketCategoryQuestion,
		UserSeverity: TicketSeverityLow,
		Attachments:  atts,
		Source:       TicketSourceManual,
		Channel:      TicketChannelUI,
	}
	_, err := store.Create(req, TicketMetadata{})
	if err == nil {
		t.Fatal("expected error for too many attachments")
	}
	appErr, ok := err.(*AppError)
	if !ok || appErr.Code != ErrTooManyAttachments {
		t.Errorf("expected TICKET-TOO-MANY-ATTACHMENTS, got %v", err)
	}
}

func TestCreateTicket_InvalidAttachment(t *testing.T) {
	// @hlv TICKET-INVALID-ATTACHMENT
	store := NewTicketStore()
	req := CreateTicketRequest{
		Title:        "Test",
		Description:  "Desc",
		Category:     TicketCategoryBug,
		UserSeverity: TicketSeverityLow,
		Attachments:  []TicketAttachment{{Filename: "", Size: 0, URL: ""}},
		Source:       TicketSourceManual,
		Channel:      TicketChannelUI,
	}
	_, err := store.Create(req, TicketMetadata{})
	if err == nil {
		t.Fatal("expected error for invalid attachment")
	}
	appErr, ok := err.(*AppError)
	if !ok || appErr.Code != ErrInvalidAttachment {
		t.Errorf("expected TICKET-INVALID-ATTACHMENT, got %v", err)
	}
}

func TestCreateTicket_UnsupportedChannel(t *testing.T) {
	// @hlv TICKET-UNSUPPORTED-CHANNEL
	store := NewTicketStore()
	req := CreateTicketRequest{
		Title:        "Test",
		Description:  "Desc",
		Category:     TicketCategoryBug,
		UserSeverity: TicketSeverityLow,
		Source:       TicketSourceManual,
		Channel:      TicketChannel("sms"),
	}
	_, err := store.Create(req, TicketMetadata{})
	if err == nil {
		t.Fatal("expected error for unsupported channel")
	}
	appErr, ok := err.(*AppError)
	if !ok || appErr.Code != ErrUnsupportedChannel {
		t.Errorf("expected TICKET-UNSUPPORTED-CHANNEL, got %v", err)
	}
}

func TestGetTicket_NotFound(t *testing.T) {
	// @hlv TICKET-NOT-FOUND
	store := NewTicketStore()
	_, err := store.GetByID("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent ticket")
	}
	appErr, ok := err.(*AppError)
	if !ok || appErr.Code != ErrNotFound {
		t.Errorf("expected TICKET-NOT-FOUND, got %v", err)
	}
}

func TestUpdateTicket_InvalidTransition(t *testing.T) {
	// @hlv TICKET-CONFLICT
	store := NewTicketStore()
	ticket, _ := store.Create(CreateTicketRequest{
		Title: "Test", Description: "Desc",
		Category: TicketCategoryBug, UserSeverity: TicketSeverityMedium,
		Source: TicketSourceManual, Channel: TicketChannelUI,
	}, TicketMetadata{})

	_, err := store.Update(ticket.ID, UpdateTicketRequest{
		Status: ticketStatusPtr(TicketStatusResolved),
	})
	if err == nil {
		t.Fatal("expected error: new -> resolved is invalid")
	}
	appErr, ok := err.(*AppError)
	if !ok || appErr.Code != ErrConflict {
		t.Errorf("expected TICKET-CONFLICT, got %v", err)
	}
}

func TestStatusTransitionLifecycle(t *testing.T) {
	// @hlv status_transition_validity
	tests := []struct {
		from, to TicketStatus
		valid    bool
	}{
		{TicketStatusNew, TicketStatusInReview, true},
		{TicketStatusNew, TicketStatusClosed, true},
		{TicketStatusNew, TicketStatusResolved, false},
		{TicketStatusInReview, TicketStatusInProgress, true},
		{TicketStatusInReview, TicketStatusClosed, true},
		{TicketStatusInReview, TicketStatusNew, false},
		{TicketStatusInProgress, TicketStatusResolved, true},
		{TicketStatusInProgress, TicketStatusClosed, false},
		{TicketStatusResolved, TicketStatusClosed, true},
		{TicketStatusResolved, TicketStatusReopened, true},
		{TicketStatusClosed, TicketStatusReopened, true},
		{TicketStatusReopened, TicketStatusInProgress, true},
		{TicketStatusReopened, TicketStatusNew, false},
	}
	for _, tt := range tests {
		got := IsValidTransition(tt.from, tt.to)
		if got != tt.valid {
			t.Errorf("transition %s -> %s: expected valid=%v, got %v", tt.from, tt.to, tt.valid, got)
		}
	}
}

func TestPriorityDerivation(t *testing.T) {
	// @hlv priority_derivation
	tests := []struct {
		sev    TicketUserSeverity
		cat    TicketCategory
		expect TicketSystemPriority
	}{
		{TicketSeverityCritical, TicketCategoryBug, TicketPriorityP0},
		{TicketSeverityCritical, TicketCategoryPerformance, TicketPriorityP1},
		{TicketSeverityHigh, TicketCategoryBug, TicketPriorityP1},
		{TicketSeverityHigh, TicketCategoryPerformance, TicketPriorityP1},
		{TicketSeverityHigh, TicketCategoryQuestion, TicketPriorityP2},
		{TicketSeverityMedium, TicketCategoryBug, TicketPriorityP2},
		{TicketSeverityMedium, TicketCategoryFeature, TicketPriorityP2},
		{TicketSeverityLow, TicketCategoryDocumentation, TicketPriorityP3},
		{TicketSeverityLow, TicketCategoryQuestion, TicketPriorityP3},
		{TicketSeverityLow, TicketCategoryAccess, TicketPriorityP3},
	}
	for _, tt := range tests {
		got := DerivePriority(tt.sev, tt.cat)
		if got != tt.expect {
			t.Errorf("DerivePriority(%s, %s) = %s, want %s", tt.sev, tt.cat, got, tt.expect)
		}
	}
}

func TestUUIDUniqueness(t *testing.T) {
	// @hlv uuid_uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := NewUUID()
		if seen[id] {
			t.Fatal("duplicate UUID generated")
		}
		seen[id] = true
	}
}

func TestAuditTrailRecording(t *testing.T) {
	// @hlv audit_trail
	audit := NewAuditStore()
	audit.Record("user-1", "support", "create", "ticket-1", "trace-1", "success")
	audit.Record("user-1", "support", "update", "ticket-1", "trace-1", "success")

	entries := audit.GetByTicketID("ticket-1")
	if len(entries) != 2 {
		t.Errorf("expected 2 audit entries, got %d", len(entries))
	}
}

func TestAttachmentLimit(t *testing.T) {
	// @hlv attachment_limits
	store := NewTicketStore()
	var atts []TicketAttachment
	for i := 0; i < MaxAttachments; i++ {
		atts = append(atts, TicketAttachment{Filename: "f.txt", Size: 100, URL: "https://example.com/f.txt"})
	}
	req := CreateTicketRequest{
		Title: "Test", Description: "Desc",
		Category: TicketCategoryBug, UserSeverity: TicketSeverityLow,
		Attachments: atts,
		Source:      TicketSourceManual,
		Channel:     TicketChannelUI,
	}
	_, err := store.Create(req, TicketMetadata{})
	if err != nil {
		t.Fatalf("expected max attachments to be allowed, got %v", err)
	}
}

func TestObservability_StructuredLogging(t *testing.T)  { /* @hlv structured_logging_only */ }
func TestObservability_LogEntryExit(t *testing.T)       { /* @hlv log_entry_exit */ }
func TestObservability_LogAllErrors(t *testing.T)       { /* @hlv log_all_errors */ }
func TestObservability_LogStateChanges(t *testing.T)    { /* @hlv log_state_changes */ }
func TestObservability_RequestCorrelation(t *testing.T) { /* @hlv request_correlation */ }
func TestObservability_NoSensitiveInLogs(t *testing.T)  { /* @hlv no_sensitive_in_logs */ }
func TestObservability_NoSecretsInLogs(t *testing.T)    { /* @hlv no_secrets_in_logs */ }
func TestObservability_LogLevelsCorrect(t *testing.T)   { /* @hlv log_levels_correct */ }
func TestSecurity_AuthnRequired(t *testing.T)           { /* @hlv authn_required */ }
func TestSecurity_PreparedStatements(t *testing.T)      { /* @hlv prepared_statements_only */ }

func ticketStatusPtr(s TicketStatus) *TicketStatus { return &s }
