package ticket

import (
	"log/slog"
	"time"
)

// @hlv audit_trail_logged all CLI operations logged in audit trail
// @hlv channel_cli channel=cli for all CLI operations
// @hlv same_ticket_model same backend and ticket model as web interface
// @hlv list_respects_visibility users see only own manual tickets
// @hlv delete_requires_guardrails delete requires MFA, env guard, typed confirmation

func Dispatch(input *TicketCliInput) *TicketCliOutput {
	if input == nil {
		// @hlv CLI_MISSING_REQUIRED_FIELD
		return errorOutput(ErrMissingRequiredField, "input is nil")
	}

	slog.Info("ticket.dispatch",
		"command", input.Command,
		"actor", input.Actor,
		"trace_id", input.TraceID,
	)

	switch input.Command {
	case "create":
		return ExecuteCreate(input)
	case "list":
		return ExecuteList(input)
	case "show":
		return ExecuteShow(input)
	case "comment":
		return ExecuteComment(input)
	case "update":
		return ExecuteUpdate(input)
	case "close":
		return ExecuteClose(input)
	case "reopen":
		return ExecuteReopen(input)
	case "delete":
		return ExecuteDelete(input)
	default:
		// @hlv CLI_INVALID_COMMAND
		return errorOutput(ErrInvalidCommand, "unsupported ticket command: "+input.Command)
	}
}

func errorOutput(code, message string) *TicketCliOutput {
	return &TicketCliOutput{
		Success: false,
		Message: message,
		Error:   &CliError{Code: code, Message: message},
	}
}

func successOutput(ticket *TicketData, message string, audit *AuditEntry) *TicketCliOutput {
	return &TicketCliOutput{
		Success:    true,
		Ticket:     ticket,
		Message:    message,
		AuditEntry: audit,
	}
}

func listOutput(tickets []TicketData, message string, audit *AuditEntry) *TicketCliOutput {
	return &TicketCliOutput{
		Success:    true,
		Tickets:    tickets,
		Message:    message,
		AuditEntry: audit,
	}
}

func buildAudit(input *TicketCliInput, ticketID string, result string) *AuditEntry {
	return &AuditEntry{
		ID:        newAuditEntryID(),
		Actor:     input.Actor,
		Role:      input.Role,
		Command:   input.Command,
		TicketID:  ticketID,
		TraceID:   input.TraceID,
		Result:    result,
		Timestamp: timeNow(),
	}
}

func timeNow() time.Time {
	return time.Now().UTC()
}
