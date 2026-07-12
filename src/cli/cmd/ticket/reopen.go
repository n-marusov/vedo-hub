// @hlv:artifact code-cli implements spec-cli-ticket-ops-001
// @ctx: Operator reopen ticket command — lifecycle transition from closed to reopened
// @hlv:sec [AUTH_BOUNDARY] — requires operator permissions

package ticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// @ctx: reopen command — transition ticket from closed to reopened state
// @hlv CLI-TICKET-NOT-FOUND
// @hlv CLI-TICKET-UNAUTHORIZED
// @hlv CLI-TICKET-INVALID-STATE
// @hlv CLI-TICKET-AUDIT-FAILED
func ExecuteReopen(apiBase string, ticketID string, token string) error {
	slog.Info("cli.ticket.reopen.enter",
		"ticket_id", ticketID,
	)

	if ticketID == "" {
		return fmt.Errorf("CLI-MISSING-REQUIRED-FIELD: ticket_id is required")
	}

	// @hlv:sec [AUTH_BOUNDARY] — token required
	if token == "" {
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: operator token required")
	}

	reqURL := fmt.Sprintf("%s/api/v1/tickets/%s/reopen", apiBase, ticketID)

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		slog.Error("cli.ticket.reopen.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("cli.ticket.reopen.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		slog.Error("cli.ticket.reopen.not_found", "ticket_id", ticketID)
		return fmt.Errorf("CLI-TICKET-NOT-FOUND: ticket %s does not exist", ticketID)
	}

	if resp.StatusCode == http.StatusForbidden {
		slog.Error("cli.ticket.reopen.unauthorized", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: insufficient operator permissions")
	}

	if resp.StatusCode == http.StatusConflict {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			slog.Error("cli.ticket.reopen.invalid_state",
				"ticket_id", ticketID,
				"error", errResp.Error,
			)
			return fmt.Errorf("CLI-TICKET-INVALID-STATE: %s", errResp.Message)
		}
		return fmt.Errorf("CLI-TICKET-INVALID-STATE: can only reopen closed or resolved tickets")
	}

	if resp.StatusCode >= 500 {
		slog.Error("cli.ticket.reopen.audit_failed", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-AUDIT-FAILED: reopen operation succeeded but audit write failed")
	}

	var result struct {
		TicketID     string `json:"ticket_id"`
		Status       string `json:"status"`
		AuditEntryID string `json:"audit_entry_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("cli.ticket.reopen.decode_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-DECODE-ERROR: %w", err)
	}

	slog.Info("cli.ticket.reopen.success",
		"ticket_id", ticketID,
		"status", result.Status,
		"audit_entry_id", result.AuditEntryID,
	)

	fmt.Printf("Ticket %s reopened successfully\n", ticketID)
	fmt.Printf("Audit entry: %s\n", result.AuditEntryID)
	return nil
}
