package ticket

import (
	"log/slog"
)

// @hlv:sec delete_requires_guardrails — delete requires MFA, env guard, typed confirmation
func ExecuteDelete(input *TicketCliInput) *TicketCliOutput {
	if input.TicketID == "" {
		// @hlv CLI_MISSING_REQUIRED_FIELD
		return errorOutput(ErrMissingRequiredField, "ticket_id is required")
	}
	if input.Actor == "" {
		input.Actor = "anonymous"
	}
	if input.Role == "" {
		input.Role = "user"
	}

	// @hlv:sec [AUTH_BOUNDARY] — guardrail enforcement for delete operation
	if !input.MFAConfirmed || !input.EnvGuard || !input.TypedConfirm {
		audit := buildAudit(input, input.TicketID, "failure")
		slog.Warn("ticket.delete.guardrail_violation",
			"ticket_id", input.TicketID,
			"actor", input.Actor,
			"mfa", input.MFAConfirmed,
			"env_guard", input.EnvGuard,
			"typed_confirm", input.TypedConfirm,
		)
		// @hlv CLI_DELETE_GUARDRAIL_VIOLATION
		return &TicketCliOutput{
			Success:    false,
			Message:    "Delete operation requires MFA confirmation, environment guard, and typed confirmation",
			AuditEntry: audit,
			Error: &CliError{
				Code:    ErrDeleteGuardrailViolation,
				Message: "Delete operation violated guardrail requirements (MFA, env guard, confirmation)",
			},
		}
	}

	_, err := defaultStore.GetByID(input.TicketID)
	if err != nil {
		// @hlv CLI_TICKET_NOT_FOUND
		return errorOutput(ErrTicketNotFound, "ticket not found: "+input.TicketID)
	}

	if err := defaultStore.Delete(input.TicketID); err != nil {
		return errorOutput(ErrTicketNotFound, "ticket not found: "+input.TicketID)
	}

	audit := buildAudit(input, input.TicketID, "success")

	slog.Info("ticket.delete.success",
		"ticket_id", input.TicketID,
		"actor", input.Actor,
	)

	return &TicketCliOutput{
		Success:    true,
		Message:    "Ticket deleted successfully",
		AuditEntry: audit,
	}
}
