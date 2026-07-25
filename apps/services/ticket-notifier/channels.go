package main

import (
	"fmt"
	"log/slog"
	"strings"
)

func handleEmailChannel(req *NotificationRequest) (recipient string, err error) {
	slog.Info("channel.email.enter",
		"ticket_id", req.TicketID,
		"event_type", req.EventType,
	)
	recipient = "user@example.com"
	if req.TemplateVars != nil {
		if name, ok := req.TemplateVars["user_name"]; ok {
			recipient = fmt.Sprintf("%s@example.com",
				strings.ToLower(strings.ReplaceAll(name, " ", ".")))
		}
	}
	slog.Info("channel.email.exit",
		"recipient", recipient,
		"status", StatusQueued,
	)
	return recipient, nil
}

func handleSlackChannel(req *NotificationRequest) (recipient string, err error) {
	slog.Info("channel.slack.enter",
		"ticket_id", req.TicketID,
		"event_type", req.EventType,
	)
	recipient = "support-team-channel"
	slog.Info("channel.slack.exit",
		"recipient", recipient,
		"status", StatusQueued,
	)
	return recipient, nil
}

func handlePagerDutyChannel(req *NotificationRequest) (recipient string, err error) {
	slog.Info("channel.pagerduty.enter",
		"ticket_id", req.TicketID,
		"event_type", req.EventType,
	)
	recipient = "pagerduty-service-id"
	slog.Info("channel.pagerduty.exit",
		"recipient", recipient,
		"status", StatusQueued,
	)
	return recipient, nil
}

func handleInAppChannel(req *NotificationRequest) (recipient string, err error) {
	slog.Info("channel.in_app.enter",
		"ticket_id", req.TicketID,
		"event_type", req.EventType,
	)
	recipient = "user-uuid-for-john-doe"
	slog.Info("channel.in_app.exit",
		"recipient", recipient,
		"status", StatusQueued,
	)
	return recipient, nil
}
