package ticket

import (
	"log/slog"
)

func ExecuteCreate(input *TicketCliInput) *TicketCliOutput {
	if input.Title == "" || input.Description == "" || input.Category == "" || input.Severity == "" {
		// @hlv CLI_MISSING_REQUIRED_FIELD
		return errorOutput(ErrMissingRequiredField, "title, description, category, and severity are required")
	}

	if input.Actor == "" {
		input.Actor = "anonymous"
	}
	if input.Role == "" {
		input.Role = "user"
	}
	if input.Channel == "" {
		input.Channel = "cli"
	}
	if input.Source == "" {
		input.Source = "manual"
	}

	ticket := defaultStore.Create(input)

	audit := buildAudit(input, ticket.ID, "success")

	slog.Info("ticket.create.success",
		"ticket_id", ticket.ID,
		"actor", input.Actor,
		"category", input.Category,
		"severity", input.Severity,
		"channel", input.Channel,
	)

	return successOutput(ticket, "Ticket created successfully", audit)
}
