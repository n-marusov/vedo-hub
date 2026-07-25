package ticket

import (
	"fmt"
	"log/slog"
)

func ExecuteList(input *TicketCliInput) *TicketCliOutput {
	if input.Actor == "" {
		input.Actor = "anonymous"
	}
	if input.Role == "" {
		input.Role = "user"
	}

	tickets := defaultStore.List(input.Actor)

	audit := buildAudit(input, "", "success")

	msg := fmt.Sprintf("Found %d tickets", len(tickets))

	slog.Info("ticket.list.success",
		"actor", input.Actor,
		"count", len(tickets),
	)

	return listOutput(tickets, msg, audit)
}
