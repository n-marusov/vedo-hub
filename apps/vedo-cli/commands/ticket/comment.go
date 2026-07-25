package ticket

import (
	"log/slog"
)

func ExecuteComment(input *TicketCliInput) *TicketCliOutput {
	if input.TicketID == "" || input.Comment == "" {
		// @hlv CLI_MISSING_REQUIRED_FIELD
		return errorOutput(ErrMissingRequiredField, "ticket_id and comment are required")
	}
	if input.Actor == "" {
		input.Actor = "anonymous"
	}
	if input.Role == "" {
		input.Role = "user"
	}

	ticket, err := defaultStore.AddComment(input.TicketID, input.Actor, input.Comment)
	if err != nil {
		// @hlv CLI_TICKET_NOT_FOUND
		return errorOutput(ErrTicketNotFound, "ticket not found: "+input.TicketID)
	}

	audit := buildAudit(input, input.TicketID, "success")

	slog.Info("ticket.comment.success",
		"ticket_id", input.TicketID,
		"actor", input.Actor,
	)

	return successOutput(ticket, "Comment added successfully", audit)
}
