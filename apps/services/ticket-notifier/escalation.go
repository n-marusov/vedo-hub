// @hlv: [ESCALATION] P0/P1 escalation routing according to escalation matrix
// @hlv All ticket lifecycle events trigger notifications
// @hlv P0/P1 tickets trigger escalation notifications
// @hlv P0/P1 escalation includes runbook links

package main

import "log/slog"

type EscalationTarget struct {
	Channel string
	Reason  string
}

var escalationMatrix = map[string][]EscalationTarget{
	"p0": {
		{Channel: "pagerduty", Reason: "P0 critical incident — immediate PagerDuty alert"},
		{Channel: "slack", Reason: "P0 critical incident — Slack notification"},
	},
	"p1": {
		{Channel: "slack", Reason: "P1 high-severity incident — Slack notification"},
	},
}

func routeEscalation(req *NotificationRequest) []NotificationRequest {
	sp := req.TicketData.SystemPriority
	targets, ok := escalationMatrix[sp]
	if !ok {
		slog.Debug("escalation.no_route",
			"ticket_id", req.TicketID,
			"system_priority", sp,
		)
		return nil
	}

	slog.Info("escalation.routing",
		"ticket_id", req.TicketID,
		"system_priority", sp,
		"targets", len(targets),
	)

	var extras []NotificationRequest
	for _, t := range targets {
		vars := make(map[string]string)
		if req.TemplateVars != nil {
			for k, v := range req.TemplateVars {
				vars[k] = v
			}
		}
		vars["escalation_reason"] = t.Reason
		vars["runbook_url"] = "https://runbooks.vedo.core/" + req.TicketData.Category + "-" + sp
		vars["severity"] = sp

		extras = append(extras, NotificationRequest{
			TicketID:            req.TicketID,
			EventType:           "escalated",
			TicketData:          req.TicketData,
			RecipientType:       "support",
			NotificationChannel: t.Channel,
			TemplateVars:        vars,
		})
	}

	return extras
}
