package ticket

import (
	"log/slog"
)

func ExecuteClose(input *TicketCliInput) *TicketCliOutput {
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

	ticket, err := defaultStore.SetStatus(input.TicketID, "closed")
	if err != nil {
		if err.Error() == ErrTicketNotFound || err.Error()[:len(ErrTicketNotFound)] == ErrTicketNotFound {
			// @hlv CLI_TICKET_NOT_FOUND
			return errorOutput(ErrTicketNotFound, "ticket not found: "+input.TicketID)
		}
		// @hlv CLI_INVALID_STATE_TRANSITION
		return errorOutput(ErrInvalidStateTransition, err.Error())
	}

	audit := buildAudit(input, input.TicketID, "success")

	slog.Info("ticket.close.success",
		"ticket_id", input.TicketID,
		"actor", input.Actor,
	)

	return successOutput(ticket, "Ticket closed successfully", audit)
}
