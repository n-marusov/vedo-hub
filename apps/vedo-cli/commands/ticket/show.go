package ticket

import (
	"log/slog"
)

func ExecuteShow(input *TicketCliInput) *TicketCliOutput {
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

	ticket, err := defaultStore.GetByID(input.TicketID)
	if err != nil {
		// @hlv CLI_TICKET_NOT_FOUND
		return errorOutput(ErrTicketNotFound, "ticket not found: "+input.TicketID)
	}

	audit := buildAudit(input, input.TicketID, "success")

	slog.Info("ticket.show.success",
		"ticket_id", input.TicketID,
		"actor", input.Actor,
	)

	return successOutput(ticket, "Ticket retrieved successfully", audit)
}
