// @hlv:artifact code-cli implements spec-cli-ticket-ops-001
// @ctx: Operator close ticket command — lifecycle transition to closed
// @hlv:sec [AUTH_BOUNDARY] — requires operator permissions, P0 closure restriction

package ticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// @ctx: close command — transition ticket to closed state
// @hlv CLI-TICKET-NOT-FOUND
// @hlv CLI-TICKET-UNAUTHORIZED
// @hlv CLI-TICKET-INVALID-STATE
// @hlv CLI-TICKET-AUDIT-FAILED
func ExecuteClose(apiBase string, ticketID string, token string) error {
	slog.Info("cli.ticket.close.enter",
		"ticket_id", ticketID,
	)

	if ticketID == "" {
		return fmt.Errorf("CLI-MISSING-REQUIRED-FIELD: ticket_id is required")
	}

	// @hlv:sec [AUTH_BOUNDARY] — token required
	if token == "" {
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: operator token required")
	}

	reqURL := fmt.Sprintf("%s/api/v1/tickets/%s/close", apiBase, ticketID)

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		slog.Error("cli.ticket.close.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("cli.ticket.close.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		slog.Error("cli.ticket.close.not_found", "ticket_id", ticketID)
		return fmt.Errorf("CLI-TICKET-NOT-FOUND: ticket %s does not exist", ticketID)
	}

	if resp.StatusCode == http.StatusForbidden {
		slog.Error("cli.ticket.close.unauthorized", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: insufficient operator permissions")
	}

	if resp.StatusCode == http.StatusConflict {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			slog.Error("cli.ticket.close.invalid_state",
				"ticket_id", ticketID,
				"error", errResp.Error,
			)
			return fmt.Errorf("CLI-TICKET-INVALID-STATE: %s", errResp.Message)
		}
		return fmt.Errorf("CLI-TICKET-INVALID-STATE: cannot close ticket in current state")
	}

	if resp.StatusCode >= 500 {
		slog.Error("cli.ticket.close.audit_failed", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-AUDIT-FAILED: close operation succeeded but audit write failed")
	}

	var result struct {
		TicketID     string `json:"ticket_id"`
		Status       string `json:"status"`
		AuditEntryID string `json:"audit_entry_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("cli.ticket.close.decode_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-DECODE-ERROR: %w", err)
	}

	slog.Info("cli.ticket.close.success",
		"ticket_id", ticketID,
		"status", result.Status,
		"audit_entry_id", result.AuditEntryID,
	)

	fmt.Printf("Ticket %s closed successfully\n", ticketID)
	fmt.Printf("Audit entry: %s\n", result.AuditEntryID)
	return nil
}
