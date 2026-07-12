// Operator show ticket command — retrieve single ticket by ID

package ticket

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// show command — get ticket details by ID
func ExecuteShow(apiBase string, ticketID string, token string) error {
	slog.Info("cli.ticket.show.enter",
		"ticket_id", ticketID,
	)

	if ticketID == "" {
		return fmt.Errorf("CLI-MISSING-REQUIRED-FIELD: ticket_id is required")
	}

	if token == "" {
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: operator token required")
	}

	reqURL := fmt.Sprintf("%s/api/v1/tickets/%s", apiBase, ticketID)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		slog.Error("cli.ticket.show.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("cli.ticket.show.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		slog.Error("cli.ticket.show.not_found", "ticket_id", ticketID)
		return fmt.Errorf("CLI-TICKET-NOT-FOUND: ticket %s does not exist", ticketID)
	}

	if resp.StatusCode == http.StatusForbidden {
		slog.Error("cli.ticket.show.unauthorized", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: insufficient operator permissions")
	}

	var ticket map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		slog.Error("cli.ticket.show.decode_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-DECODE-ERROR: %w", err)
	}

	slog.Info("cli.ticket.show.success",
		"ticket_id", ticketID,
		"status", ticket["status"],
	)

	renderTicketDetail(ticket)
	return nil
}

func renderTicketDetail(ticket map[string]interface{}) {
	fmt.Printf("ID:               %s\n", ticket["id"])
	fmt.Printf("Status:           %s\n", ticket["status"])
	fmt.Printf("Priority:         %s\n", ticket["system_priority"])
	fmt.Printf("Source:           %s\n", ticket["source"])
	fmt.Printf("Channel:          %s\n", ticket["channel"])
	fmt.Printf("Title:            %s\n", ticket["title"])
	fmt.Printf("Category:         %s\n", ticket["category"])
	fmt.Printf("Severity:         %s\n", ticket["user_severity"])
	fmt.Printf("Created:          %s\n", ticket["created_at"])
	fmt.Printf("Updated:          %s\n", ticket["updated_at"])

	if desc, ok := ticket["description"].(string); ok && desc != "" {
		fmt.Printf("\nDescription:\n%s\n", desc)
	}

	if steps, ok := ticket["steps_to_reproduce"].(string); ok && steps != "" {
		fmt.Printf("\nSteps to Reproduce:\n%s\n", steps)
	}

	if comments, ok := ticket["comments"].([]interface{}); ok && len(comments) > 0 {
		fmt.Printf("\nComments (%d):\n", len(comments))
		for _, c := range comments {
			cm := c.(map[string]interface{})
			fmt.Printf("  [%s] %s: %s\n", cm["created_at"], cm["author"], cm["text"])
		}
	}
}
