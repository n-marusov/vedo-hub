// @ctx: domain types and error codes for ticket-notifier — maps TICKET-NOTIFY-001 contract

package main

import (
	"crypto/rand"
	"fmt"
	"time"
)

type NotificationRequest struct {
	TicketID            string            `json:"ticket_id" binding:"required"`
	EventType           string            `json:"event_type" binding:"required"`
	TicketData          TicketData        `json:"ticket_data" binding:"required"`
	RecipientType       string            `json:"recipient_type" binding:"required"`
	NotificationChannel string            `json:"notification_channel" binding:"required"`
	TemplateVars        map[string]string `json:"template_vars"`
}

type TicketData struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	UserSeverity   string `json:"user_severity"`
	SystemPriority string `json:"system_priority"`
	Source         string `json:"source"`
	Channel        string `json:"channel"`
	Status         string `json:"status"`
}

type NotificationResponse struct {
	NotificationID string  `json:"notification_id"`
	Status         string  `json:"status"`
	Channel        string  `json:"channel"`
	Recipient      string  `json:"recipient"`
	Timestamp      string  `json:"timestamp"`
	Error          *string `json:"error"`
	RetryCount     int     `json:"retry_count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Ref     string `json:"ticket_management_system_ref"`
}

var ValidEventTypes = map[string]bool{
	"created":        true,
	"status_changed": true,
	"comment_added":  true,
	"escalated":      true,
}

var ValidChannels = map[string]bool{
	"email":     true,
	"slack":     true,
	"pagerduty": true,
	"in_app":    true,
}

var ValidRecipientTypes = map[string]bool{
	"user":      true,
	"support":   true,
	"admin":     true,
	"pagerduty": true,
	"slack":     true,
}

const (
	ErrInvalidRequest     = "NOTIFY-INVALID-REQUEST"
	ErrUnauthorized       = "NOTIFY-UNAUTHORIZED"
	ErrChannelUnavailable = "NOTIFY-CHANNEL-UNAVAILABLE"
	ErrTemplateError      = "NOTIFY-TEMPLATE-ERROR"
	ErrRateLimited        = "NOTIFY-RATE-LIMITED"
	ErrRecipientInvalid   = "NOTIFY-RECIPIENT-INVALID"
)

const (
	StatusQueued    = "queued"
	StatusSent      = "sent"
	StatusDelivered = "delivered"
	StatusFailed    = "failed"
)

type AppError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Ref     string `json:"ticket_management_system_ref,omitempty"`
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}

func NewUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("rand.Read failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func IsValidUUID(uuid string) bool {
	if len(uuid) != 36 {
		return false
	}
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		return false
	}
	return true
}

func timestampNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
