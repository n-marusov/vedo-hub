// Operator ticket list command — filter by source, status, category, priority

package ticket

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// list command — operator can view both manual and telemetry tickets
func ExecuteList(apiBase string, source []string, status []string, category []string, priority []string, token string) error {
	slog.Info("cli.ticket.list.enter",
		"source", source,
		"status", status,
		"category", category,
		"priority", priority,
	)

	if token == "" {
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: operator token required")
	}

	// build query params
	params := url.Values{}
	for _, s := range source {
		params.Add("source", s)
	}
	for _, s := range status {
		params.Add("status", s)
	}
	for _, c := range category {
		params.Add("category", c)
	}
	for _, p := range priority {
		params.Add("priority", p)
	}

	reqURL := fmt.Sprintf("%s/api/v1/tickets?%s", apiBase, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		slog.Error("cli.ticket.list.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("cli.ticket.list.request_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-NETWORK-ERROR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		slog.Error("cli.ticket.list.unauthorized", "status", resp.StatusCode)
		return fmt.Errorf("CLI-TICKET-UNAUTHORIZED: insufficient operator permissions")
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.NewDecoder(resp.Body).Decode(&errResp) == nil {
			return fmt.Errorf("CLI-TICKET-INVALID-FILTER: %s", errResp.Message)
		}
		return fmt.Errorf("CLI-TICKET-INVALID-FILTER: invalid filter values")
	}

	var result struct {
		Tickets []map[string]interface{} `json:"tickets"`
		Total   int                      `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("cli.ticket.list.decode_failed", "error", err)
		return fmt.Errorf("CLI-TICKET-DECODE-ERROR: %w", err)
	}

	slog.Info("cli.ticket.list.success",
		"count", len(result.Tickets),
		"total", result.Total,
	)

	// render output
	renderTicketList(result.Tickets)
	return nil
}

func renderTicketList(tickets []map[string]interface{}) {
	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
		return
	}

	fmt.Printf("%-8s %-12s %-10s %-8s %-15s %s\n", "ID", "STATUS", "PRIORITY", "SOURCE", "CATEGORY", "TITLE")
	fmt.Println(strings.Repeat("-", 100))

	for _, t := range tickets {
		id := truncateStr(fmt.Sprintf("%v", t["id"]), 8)
		status := fmt.Sprintf("%v", t["status"])
		priority := fmt.Sprintf("%v", t["system_priority"])
		source := fmt.Sprintf("%v", t["source"])
		category := fmt.Sprintf("%v", t["category"])
		title := truncateStr(fmt.Sprintf("%v", t["title"]), 40)
		fmt.Printf("%-8s %-12s %-10s %-8s %-15s %s\n", id, status, priority, source, category, title)
	}
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
