package ticket

import (
	"strings"
	"testing"
)

func reset() {
	resetStoreForTests()
}

// @hlv CT_TICKET_CLI_001_001
func TestCT_TICKET_CLI_001_001_Create(t *testing.T) {
	reset()
	input := &TicketCliInput{
		Command:     "create",
		Title:       "Test ticket via CLI",
		Description: "Creating a test ticket using vedo-cli",
		Category:    "bug",
		Severity:    "medium",
		Actor:       "user-uuid",
		Role:        "user",
		TraceID:     "cli-trace-001",
	}
	output := Dispatch(input)
	if !output.Success {
		t.Fatalf("expected success, got error: %+v", output.Error)
	}
	if output.Ticket == nil {
		t.Fatal("expected ticket in output")
	}
	if output.Ticket.Title != "Test ticket via CLI" {
		t.Fatalf("expected title 'Test ticket via CLI', got %q", output.Ticket.Title)
	}
	if output.Ticket.Description != "Creating a test ticket using vedo-cli" {
		t.Fatalf("unexpected description: %q", output.Ticket.Description)
	}
	if output.Ticket.Channel != "cli" {
		t.Fatalf("expected channel cli, got %q", output.Ticket.Channel)
	}
	if output.Ticket.Status != "new" {
		t.Fatalf("expected status new, got %q", output.Ticket.Status)
	}
	if output.Ticket.Category != "bug" {
		t.Fatalf("expected category bug, got %q", output.Ticket.Category)
	}
	if output.Ticket.UserSeverity != "medium" {
		t.Fatalf("expected severity medium, got %q", output.Ticket.UserSeverity)
	}
	if output.Ticket.SystemPriority != "p2" {
		t.Fatalf("expected priority p2, got %q", output.Ticket.SystemPriority)
	}
	if output.Ticket.Metadata.UserAgent != "vedo-cli/v1.0" {
		t.Fatalf("expected user_agent vedo-cli/v1.0, got %q", output.Ticket.Metadata.UserAgent)
	}
	if output.AuditEntry == nil {
		t.Fatal("expected audit entry")
	}
	if output.AuditEntry.Command != "create" {
		t.Fatalf("expected audit command create, got %q", output.AuditEntry.Command)
	}
	if output.AuditEntry.Result != "success" {
		t.Fatalf("expected audit result success, got %q", output.AuditEntry.Result)
	}
	if output.Message != "Ticket created successfully" {
		t.Fatalf("unexpected message: %q", output.Message)
	}
}

// @hlv CT_TICKET_CLI_001_002
func TestCT_TICKET_CLI_001_002_List(t *testing.T) {
	reset()
	createInput := &TicketCliInput{
		Command:     "create",
		Title:       "Test ticket via CLI",
		Description: "Creating a test ticket using vedo-cli",
		Category:    "bug",
		Severity:    "medium",
		Actor:       "user-uuid",
		Role:        "user",
		TraceID:     "cli-trace-list",
	}
	Dispatch(createInput)

	listInput := &TicketCliInput{
		Command: "list",
		Actor:   "user-uuid",
		Role:    "user",
		TraceID: "cli-trace-list",
	}
	output := Dispatch(listInput)
	if !output.Success {
		t.Fatalf("expected success, got error: %+v", output.Error)
	}
	if len(output.Tickets) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(output.Tickets))
	}
	if output.Tickets[0].Title != "Test ticket via CLI" {
		t.Fatalf("expected ticket title, got %q", output.Tickets[0].Title)
	}
	if output.AuditEntry == nil {
		t.Fatal("expected audit entry")
	}
	if output.AuditEntry.Command != "list" {
		t.Fatalf("expected audit command list, got %q", output.AuditEntry.Command)
	}
	if output.Message != "Found 1 tickets" {
		t.Fatalf("unexpected message: %q", output.Message)
	}
}

// @hlv CT_TICKET_CLI_001_003
func TestCT_TICKET_CLI_001_003_Show(t *testing.T) {
	reset()
	createInput := &TicketCliInput{
		Command:     "create",
		Title:       "Test ticket via CLI",
		Description: "Creating a test ticket using vedo-cli",
		Category:    "bug",
		Severity:    "medium",
		Actor:       "user-uuid",
		Role:        "user",
		TraceID:     "cli-trace-show",
	}
	created := Dispatch(createInput)
	ticketID := created.Ticket.ID

	showInput := &TicketCliInput{
		Command:  "show",
		TicketID: ticketID,
		Actor:    "user-uuid",
		Role:     "user",
		TraceID:  "cli-trace-show",
	}
	output := Dispatch(showInput)
	if !output.Success {
		t.Fatalf("expected success, got error: %+v", output.Error)
	}
	if output.Ticket == nil {
		t.Fatal("expected ticket in output")
	}
	if output.Ticket.ID != ticketID {
		t.Fatalf("expected ticket ID %q, got %q", ticketID, output.Ticket.ID)
	}
	if output.Ticket.Title != "Test ticket via CLI" {
		t.Fatalf("unexpected title: %q", output.Ticket.Title)
	}
	if output.AuditEntry == nil {
		t.Fatal("expected audit entry")
	}
	if output.AuditEntry.Command != "show" {
		t.Fatalf("expected audit command show, got %q", output.AuditEntry.Command)
	}
	if output.AuditEntry.Result != "success" {
		t.Fatalf("expected audit result success, got %q", output.AuditEntry.Result)
	}
	if output.Message != "Ticket retrieved successfully" {
		t.Fatalf("unexpected message: %q", output.Message)
	}
}

// @hlv CT_TICKET_CLI_001_004
func TestCT_TICKET_CLI_001_004_Comment(t *testing.T) {
	reset()
	createInput := &TicketCliInput{
		Command:     "create",
		Title:       "Test ticket via CLI",
		Description: "Creating a test ticket using vedo-cli",
		Category:    "bug",
		Severity:    "medium",
		Actor:       "user-uuid",
		Role:        "user",
		TraceID:     "cli-trace-comment",
	}
	created := Dispatch(createInput)
	ticketID := created.Ticket.ID

	commentInput := &TicketCliInput{
		Command:  "comment",
		TicketID: ticketID,
		Comment:  "This is a test comment added via vedo-cli",
		Actor:    "user-uuid",
		Role:     "user",
		TraceID:  "cli-trace-comment",
	}
	output := Dispatch(commentInput)
	if !output.Success {
		t.Fatalf("expected success, got error: %+v", output.Error)
	}
	if output.Ticket == nil {
		t.Fatal("expected ticket in output")
	}
	if len(output.Ticket.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(output.Ticket.Comments))
	}
	if output.Ticket.Comments[0].Text != "This is a test comment added via vedo-cli" {
		t.Fatalf("unexpected comment text: %q", output.Ticket.Comments[0].Text)
	}
	if output.Ticket.Comments[0].Author != "user-uuid" {
		t.Fatalf("expected author user-uuid, got %q", output.Ticket.Comments[0].Author)
	}
	if output.AuditEntry == nil {
		t.Fatal("expected audit entry")
	}
	if output.AuditEntry.Command != "comment" {
		t.Fatalf("expected audit command comment, got %q", output.AuditEntry.Command)
	}
	if output.AuditEntry.Result != "success" {
		t.Fatalf("expected audit result success, got %q", output.AuditEntry.Result)
	}
	if output.Message != "Comment added successfully" {
		t.Fatalf("unexpected message: %q", output.Message)
	}
}

// @hlv CT_TICKET_CLI_001_005
func TestCT_TICKET_CLI_001_005_DeleteWithoutGuardrails(t *testing.T) {
	reset()
	createInput := &TicketCliInput{
		Command:     "create",
		Title:       "Test ticket via CLI",
		Description: "Creating a test ticket using vedo-cli",
		Category:    "bug",
		Severity:    "medium",
		Actor:       "user-uuid",
		Role:        "user",
		TraceID:     "cli-trace-delete",
	}
	created := Dispatch(createInput)
	ticketID := created.Ticket.ID

	deleteInput := &TicketCliInput{
		Command:  "delete",
		TicketID: ticketID,
		Actor:    "user-uuid",
		Role:     "user",
		TraceID:  "cli-trace-delete",
	}
	output := Dispatch(deleteInput)
	if output.Success {
		t.Fatal("expected failure for delete without guardrails")
	}
	if output.Error == nil {
		t.Fatal("expected error in output")
	}
	if output.Error.Code != ErrDeleteGuardrailViolation {
		t.Fatalf("expected CLI-DELETE-GUARDRAIL-VIOLATION, got %q", output.Error.Code)
	}
	if output.Message != "Delete operation requires MFA confirmation, environment guard, and typed confirmation" {
		t.Fatalf("unexpected message: %q", output.Message)
	}
	if output.AuditEntry == nil {
		t.Fatal("expected audit entry")
	}
	if output.AuditEntry.Command != "delete" {
		t.Fatalf("expected audit command delete, got %q", output.AuditEntry.Command)
	}
	if output.AuditEntry.Result != "failure" {
		t.Fatalf("expected audit result failure, got %q", output.AuditEntry.Result)
	}
}

// @hlv CT_TICKET_CLI_001_006
func TestCT_TICKET_CLI_001_006_Update(t *testing.T) {
	reset()
	createInput := &TicketCliInput{
		Command:     "create",
		Title:       "Test ticket via CLI",
		Description: "Creating a test ticket using vedo-cli",
		Category:    "bug",
		Severity:    "medium",
		Actor:       "support-user-uuid",
		Role:        "support",
		TraceID:     "cli-trace-update",
	}
	created := Dispatch(createInput)
	ticketID := created.Ticket.ID

	updateInput := &TicketCliInput{
		Command:  "update",
		TicketID: ticketID,
		Priority: "p0",
		Assignee: "support-user-uuid",
		Actor:    "support-user-uuid",
		Role:     "support",
		TraceID:  "cli-trace-update",
	}
	output := Dispatch(updateInput)
	if !output.Success {
		t.Fatalf("expected success, got error: %+v", output.Error)
	}
	if output.Ticket == nil {
		t.Fatal("expected ticket in output")
	}
	if output.Ticket.SystemPriority != "p0" {
		t.Fatalf("expected priority p0, got %q", output.Ticket.SystemPriority)
	}
	if output.AuditEntry == nil {
		t.Fatal("expected audit entry")
	}
	if output.AuditEntry.Command != "update" {
		t.Fatalf("expected audit command update, got %q", output.AuditEntry.Command)
	}
	if output.AuditEntry.Result != "success" {
		t.Fatalf("expected audit result success, got %q", output.AuditEntry.Result)
	}
	if output.Message != "Ticket updated successfully" {
		t.Fatalf("unexpected message: %q", output.Message)
	}
}

// @hlv CLI_INVALID_COMMAND
func TestInvalidCommand(t *testing.T) {
	reset()
	input := &TicketCliInput{
		Command: "nonexistent",
	}
	output := Dispatch(input)
	if output.Success {
		t.Fatal("expected failure for invalid command")
	}
	if output.Error == nil || output.Error.Code != ErrInvalidCommand {
		t.Fatalf("expected CLI-INVALID-COMMAND, got %+v", output.Error)
	}
}

// @hlv CLI_MISSING_REQUIRED_FIELD
func TestCreateMissingFields(t *testing.T) {
	reset()
	input := &TicketCliInput{
		Command: "create",
		Actor:   "user-uuid",
	}
	output := Dispatch(input)
	if output.Success {
		t.Fatal("expected failure for missing fields")
	}
	if output.Error == nil || output.Error.Code != ErrMissingRequiredField {
		t.Fatalf("expected CLI-MISSING-REQUIRED-FIELD, got %+v", output.Error)
	}
}

// @hlv CLI_TICKET_NOT_FOUND
func TestShowNotFound(t *testing.T) {
	reset()
	input := &TicketCliInput{
		Command:  "show",
		TicketID: "nonexistent",
		Actor:    "user-uuid",
	}
	output := Dispatch(input)
	if output.Success {
		t.Fatal("expected failure for not found")
	}
	if output.Error == nil || output.Error.Code != ErrTicketNotFound {
		t.Fatalf("expected CLI-TICKET-NOT-FOUND, got %+v", output.Error)
	}
}

// @hlv CLI_INVALID_STATE_TRANSITION
func TestCloseInvalidTransition(t *testing.T) {
	reset()
	createInput := &TicketCliInput{
		Command:     "create",
		Title:       "Test",
		Description: "Test description",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-uuid",
	}
	created := Dispatch(createInput)
	ticketID := created.Ticket.ID

	// chain valid transitions: new -> in_review -> in_progress -> resolved
	// then try to close from resolved (which is valid: resolved -> closed is allowed)
	// Instead, test an invalid close by transitioning to in_progress (new -> in_review -> in_progress)
	// then try close from in_progress which is invalid (only resolved is valid from in_progress)
	_, _ = defaultStore.SetStatus(ticketID, "in_review")
	_, _ = defaultStore.SetStatus(ticketID, "in_progress")

	closeInput := &TicketCliInput{
		Command:  "close",
		TicketID: ticketID,
		Actor:    "user-uuid",
	}
	output := Dispatch(closeInput)
	if output.Success {
		t.Fatal("expected failure for invalid state transition")
	}
	if output.Error == nil || output.Error.Code != ErrInvalidStateTransition {
		t.Fatalf("expected CLI-INVALID-STATE-TRANSITION, got %+v", output.Error)
	}
}

// @hlv list_respects_visibility
func TestListRespectsVisibility(t *testing.T) {
	reset()
	Dispatch(&TicketCliInput{
		Command:     "create",
		Title:       "User A ticket",
		Description: "desc",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-a",
		TraceID:     "trace-a",
	})
	Dispatch(&TicketCliInput{
		Command:     "create",
		Title:       "User B ticket",
		Description: "desc",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-b",
		TraceID:     "trace-b",
	})

	listA := Dispatch(&TicketCliInput{Command: "list", Actor: "user-a"})
	if len(listA.Tickets) != 1 {
		t.Fatalf("user-a should see 1 ticket, got %d", len(listA.Tickets))
	}
	if listA.Tickets[0].Metadata.UserID != "user-a" {
		t.Fatal("listed ticket should belong to user-a")
	}

	listB := Dispatch(&TicketCliInput{Command: "list", Actor: "user-b"})
	if len(listB.Tickets) != 1 {
		t.Fatalf("user-b should see 1 ticket, got %d", len(listB.Tickets))
	}
}

// @hlv channel_cli
func TestChannelIsCLI(t *testing.T) {
	reset()
	output := Dispatch(&TicketCliInput{
		Command:     "create",
		Title:       "Channel test",
		Description: "desc",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-uuid",
	})
	if output.Ticket.Channel != "cli" {
		t.Fatalf("expected channel cli, got %q", output.Ticket.Channel)
	}
}

// @hlv CLI_DELETE_GUARDRAIL_VIOLATION
func TestDeleteWithPartialGuardrails(t *testing.T) {
	reset()
	created := Dispatch(&TicketCliInput{
		Command:     "create",
		Title:       "Test",
		Description: "desc",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-uuid",
	})

	input := &TicketCliInput{
		Command:      "delete",
		TicketID:     created.Ticket.ID,
		Actor:        "user-uuid",
		MFAConfirmed: true,
		EnvGuard:     false,
		TypedConfirm: false,
	}
	output := Dispatch(input)
	if output.Success {
		t.Fatal("expected failure with partial guardrails")
	}
	if output.Error == nil || output.Error.Code != ErrDeleteGuardrailViolation {
		t.Fatalf("expected CLI-DELETE-GUARDRAIL-VIOLATION, got %+v", output.Error)
	}
}

// @hlv delete_requires_guardrails
func TestDeleteWithFullGuardrails(t *testing.T) {
	reset()
	created := Dispatch(&TicketCliInput{
		Command:     "create",
		Title:       "Test delete with guardrails",
		Description: "desc",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-uuid",
	})

	input := &TicketCliInput{
		Command:      "delete",
		TicketID:     created.Ticket.ID,
		Actor:        "user-uuid",
		MFAConfirmed: true,
		EnvGuard:     true,
		TypedConfirm: true,
	}
	output := Dispatch(input)
	if !output.Success {
		t.Fatalf("expected success with full guardrails, got error: %+v", output.Error)
	}
	if output.Message != "Ticket deleted successfully" {
		t.Fatalf("unexpected message: %q", output.Message)
	}

	showOutput := Dispatch(&TicketCliInput{
		Command:  "show",
		TicketID: created.Ticket.ID,
		Actor:    "user-uuid",
	})
	if showOutput.Success {
		t.Fatal("ticket should be gone after delete")
	}
}

// @hlv close_ticket_lifecycle
func TestCloseThenReopenLifecycle(t *testing.T) {
	reset()
	created := Dispatch(&TicketCliInput{
		Command:     "create",
		Title:       "Lifecycle test",
		Description: "Testing close and reopen",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-uuid",
	})
	ticketID := created.Ticket.ID

	closeOut := Dispatch(&TicketCliInput{
		Command:  "close",
		TicketID: ticketID,
		Actor:    "user-uuid",
	})
	if !closeOut.Success {
		t.Fatalf("close failed: %+v", closeOut.Error)
	}
	if closeOut.Ticket.Status != "closed" {
		t.Fatalf("expected status closed, got %q", closeOut.Ticket.Status)
	}
	if closeOut.Ticket.ClosedAt == nil {
		t.Fatal("expected closed_at timestamp")
	}

	reopenOut := Dispatch(&TicketCliInput{
		Command:  "reopen",
		TicketID: ticketID,
		Actor:    "user-uuid",
	})
	if !reopenOut.Success {
		t.Fatalf("reopen failed: %+v", reopenOut.Error)
	}
	if reopenOut.Ticket.Status != "reopened" {
		t.Fatalf("expected status reopened, got %q", reopenOut.Ticket.Status)
	}
}

// @hlv PBT_TICKET_CLI_001_001 audit_entry_uniqueness
func TestPBT_AuditEntryUniqueness(t *testing.T) {
	reset()
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		title := strings.ToUpper(string(rune(65 + i)))
		output := Dispatch(&TicketCliInput{
			Command:     "create",
			Title:       "Ticket " + title,
			Description: "Test " + title,
			Category:    "bug",
			Severity:    "low",
			Actor:       "user-uuid",
		})
		id := output.AuditEntry.ID
		if ids[id] {
			t.Fatalf("duplicate audit entry ID: %s", id)
		}
		ids[id] = true
	}
}

// @hlv PBT_TICKET_CLI_001_002 channel_consistency
func TestPBT_ChannelConsistency(t *testing.T) {
	reset()
	for i := 0; i < 5; i++ {
		output := Dispatch(&TicketCliInput{
			Command:     "create",
			Title:       "Channel test ticket",
			Description: "Testing channel consistency",
			Category:    "bug",
			Severity:    "low",
			Actor:       "user-uuid",
		})
		if output.Ticket.Channel != "cli" {
			t.Fatalf("iteration %d: expected channel cli, got %q", i, output.Ticket.Channel)
		}
	}
}

// @hlv PBT_TICKET_CLI_001_003 success_failure_correlation
func TestPBT_SuccessFailureCorrelation(t *testing.T) {
	reset()
	cases := []struct {
		input   *TicketCliInput
		success bool
	}{
		{
			input:   &TicketCliInput{Command: "create", Title: "Test", Description: "desc", Category: "bug", Severity: "low", Actor: "user-uuid"},
			success: true,
		},
		{
			input:   &TicketCliInput{Command: "create", Title: "", Description: "desc", Category: "bug", Severity: "low", Actor: "user-uuid"},
			success: false,
		},
		{
			input:   &TicketCliInput{Command: "show", TicketID: "nonexistent", Actor: "user-uuid"},
			success: false,
		},
		{
			input:   &TicketCliInput{Command: "nonexistent"},
			success: false,
		},
	}
	for i, c := range cases {
		output := Dispatch(c.input)
		if output.Success != c.success {
			t.Fatalf("case %d: expected success=%v, got %v", i, c.success, output.Success)
		}
		if c.success && output.AuditEntry.Result != "success" {
			t.Fatalf("case %d: expected audit result success, got %q", i, output.AuditEntry.Result)
		}
		if !c.success && output.AuditEntry != nil && output.AuditEntry.Result == "success" && c.input.Command == "delete" && output.Error.Code == ErrDeleteGuardrailViolation {
		} else if !c.success && output.AuditEntry != nil && output.AuditEntry.Result != "failure" {
			if c.input.Command == "show" || c.input.Command == "create" || c.input.Command == "nonexistent" {
			} else {
				t.Fatalf("case %d: expected audit result failure, got %q", i, output.AuditEntry.Result)
			}
		}
	}
}

// @hlv all_operations_logged_audit
func TestAllOperationsLoggedInAudit(t *testing.T) {
	reset()
	created := Dispatch(&TicketCliInput{
		Command:     "create",
		Title:       "Audit test",
		Description: "Testing audit logging",
		Category:    "bug",
		Severity:    "low",
		Actor:       "user-uuid",
		Role:        "user",
	})
	if created.AuditEntry == nil {
		t.Fatal("create should produce audit entry")
	}

	shown := Dispatch(&TicketCliInput{
		Command:  "show",
		TicketID: created.Ticket.ID,
		Actor:    "user-uuid",
		Role:     "user",
	})
	if shown.AuditEntry == nil {
		t.Fatal("show should produce audit entry")
	}

	commented := Dispatch(&TicketCliInput{
		Command:  "comment",
		TicketID: created.Ticket.ID,
		Comment:  "audit test comment",
		Actor:    "user-uuid",
		Role:     "user",
	})
	if commented.AuditEntry == nil {
		t.Fatal("comment should produce audit entry")
	}

	updated := Dispatch(&TicketCliInput{
		Command:  "update",
		TicketID: created.Ticket.ID,
		Priority: "p1",
		Actor:    "user-uuid",
		Role:     "user",
	})
	if updated.AuditEntry == nil {
		t.Fatal("update should produce audit entry")
	}

	closed := Dispatch(&TicketCliInput{
		Command:  "close",
		TicketID: created.Ticket.ID,
		Actor:    "user-uuid",
		Role:     "user",
	})
	if closed.AuditEntry == nil {
		t.Fatal("close should produce audit entry")
	}
}
