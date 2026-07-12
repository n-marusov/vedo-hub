// @hlv:artifact code-cli implements spec-cli-ticket-ops-001
// @ctx: Operator comment command — add comment to ticket
// @hlv:sec [AUTH_BOUNDARY] — requires operator permissions

package ticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// @ctx: comment command — add comment to ticket, creates audit entry
// @hlv CLI-TICKET-NOT-FOUND
// @hlv CLI-TICKET-UNAUTHORIZED
// @hlv CLI-TICKET-AUDIT-FAILED
func ExecuteComment(apiBase string, ticketID string, commentText string, token string) error {
	slog.Info("cli.ticket.comment.enter",
		"ticket_id", ticketID,
		"comment_length", len(commentText),
	)

	if ticketID == "" {
		return fmt.Errorf("CLI-MISSING-REQUIRED-FIELD: ticket_id is required")
	}
	if commentText == "" {
		return fmt.Errorf("CLI-MISSING-REQUIRED-FIELD: comment text is required")
	}

	// @hlv:sec [AUTH_BOUNDARY] — token required
	if token == "" {
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: operator token required")
	}

	payload := map[string]string{
		"text": commentText,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("cli.ticket.comment.encode_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-ENCODE-ERROR: %w", err)
	}

	reqURL := fmt.Sprintf("%s/api/v1/tickets/%s/comments", apiBase, ticketID)

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		slog.Error("cli.ticket.comment.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("cli.ticket.comment.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		slog.Error("cli.ticket.comment.not_found", "ticket_id", ticketID)
		return fmt.Errorf("CLI-TICKET-NOT-FOUND: ticket %s does not exist", ticketID)
	}

	if resp.StatusCode == http.StatusForbidden {
		slog.Error("cli.ticket.comment.unauthorized", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: insufficient operator permissions")
	}

	if resp.StatusCode >= 500 {
		slog.Error("cli.ticket.comment.audit_failed", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-AUDIT-FAILED: comment operation succeeded but audit write failed")
	}

	var result struct {
		AuditEntryID string `json:"audit_entry_id"`
		Comment      struct {
			Text      string `json:"text"`
			CreatedAt string `json:"created_at"`
		} `json:"comment"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("cli.ticket.comment.decode_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-DECODE-ERROR: %w", err)
	}

	slog.Info("cli.ticket.comment.success",
		"ticket_id", ticketID,
		"audit_entry_id", result.AuditEntryID,
	)

	fmt.Printf("Comment added to ticket %s\n", ticketID)
	fmt.Printf("Audit entry: %s\n", result.AuditEntryID)
	return nil
}
