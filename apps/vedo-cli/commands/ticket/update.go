package ticket

import (
	"log/slog"
)

func ExecuteUpdate(input *TicketCliInput) *TicketCliOutput {
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

	_, err := defaultStore.GetByID(input.TicketID)
	if err != nil {
		// @hlv CLI_TICKET_NOT_FOUND
		return errorOutput(ErrTicketNotFound, "ticket not found: "+input.TicketID)
	}

	ticket, err := defaultStore.Update(input.TicketID, input.Priority, input.Assignee)
	if err != nil {
		return errorOutput(ErrTicketNotFound, "ticket not found: "+input.TicketID)
	}

	audit := buildAudit(input, input.TicketID, "success")

	slog.Info("ticket.update.success",
		"ticket_id", input.TicketID,
		"actor", input.Actor,
		"priority", input.Priority,
	)

	return successOutput(ticket, "Ticket updated successfully", audit)
}
