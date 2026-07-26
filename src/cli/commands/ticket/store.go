package ticket

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type TicketStore struct {
	mu      sync.RWMutex
	tickets map[string]*TicketData
	counter int
}

var defaultStore = NewTicketStore()

func NewTicketStore() *TicketStore {
	return &TicketStore{
		tickets: make(map[string]*TicketData),
	}
}

func (s *TicketStore) getNextID() string {
	s.counter++
	return fmt.Sprintf("ticket-%d-%d", time.Now().UnixNano(), s.counter)
}

func (s *TicketStore) Create(input *TicketCliInput) *TicketData {
	now := time.Now().UTC()
	traceID := input.TraceID
	userAgent := "vedo-cli/v1.0"

	t := &TicketData{
		ID:     s.getNextID(),
		Source: "manual",
		Channel: func() string {
			if input.Channel != "" {
				return input.Channel
			}
			return "cli"
		}(),
		Status:       "new",
		Title:        input.Title,
		Description:  input.Description,
		Category:     input.Category,
		UserSeverity: input.Severity,
		SystemPriority: derivePriority(input.Severity, input.Category),
		Metadata: TicketMetadata{
			VedoVersion: "1.0.0",
			Environment: "production",
			UserID:      input.Actor,
			TraceID:     &traceID,
			UserAgent:   userAgent,
		},
		Attachments: make([]interface{}, 0),
		Labels:      make([]string, 0),
		Comments:    make([]TicketCommentData, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	s.tickets[t.ID] = t
	s.mu.Unlock()

	slog.Info("ticket.store.created",
		"ticket_id", t.ID,
		"channel", t.Channel,
		"category", t.Category,
		"priority", t.SystemPriority,
	)

	return t
}

func (s *TicketStore) GetByID(id string) (*TicketData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tickets[id]
	if !ok {
		return nil, errors.New(ErrTicketNotFound)
	}
	return t, nil
}

func (s *TicketStore) List(actor string) []TicketData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []TicketData
	for _, t := range s.tickets {
		if t.Metadata.UserID == actor && t.Source == "manual" {
			result = append(result, *t)
		}
	}
	return result
}

func (s *TicketStore) AddComment(ticketID, author, text string) (*TicketData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tickets[ticketID]
	if !ok {
		return nil, errors.New(ErrTicketNotFound)
	}

	comment := TicketCommentData{
		Author:    author,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}
	t.Comments = append(t.Comments, comment)
	t.UpdatedAt = time.Now().UTC()

	slog.Info("ticket.store.comment_added",
		"ticket_id", ticketID,
		"author", author,
	)

	return t, nil
}

func (s *TicketStore) Update(id, priority, assignee string) (*TicketData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tickets[id]
	if !ok {
		return nil, errors.New(ErrTicketNotFound)
	}

	if priority != "" {
		t.SystemPriority = priority
	}
	if assignee != "" {
		t.Metadata.UserID = assignee
	}
	t.UpdatedAt = time.Now().UTC()

	slog.Info("ticket.store.updated",
		"ticket_id", id,
		"priority", t.SystemPriority,
	)

	return t, nil
}

func (s *TicketStore) SetStatus(id string, status string) (*TicketData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tickets[id]
	if !ok {
		return nil, errors.New(ErrTicketNotFound)
	}

	if !isValidTransition(t.Status, status) {
		// @hlv CLI_INVALID_STATE_TRANSITION
		return nil, fmt.Errorf("%s: cannot transition from %s to %s", ErrInvalidStateTransition, t.Status, status)
	}

	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	if status == "closed" {
		now := time.Now().UTC()
		t.ClosedAt = &now
	}

	slog.Info("ticket.store.status_changed",
		"ticket_id", id,
		"status", status,
	)

	return t, nil
}

func (s *TicketStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tickets[id]; !ok {
		return errors.New(ErrTicketNotFound)
	}

	delete(s.tickets, id)

	slog.Info("ticket.store.deleted",
		"ticket_id", id,
	)

	return nil
}

func derivePriority(severity, category string) string {
	switch severity {
	case "critical":
		if category == "bug" {
			return "p0"
		}
		return "p1"
	case "high":
		if category == "bug" || category == "performance" {
			return "p1"
		}
		return "p2"
	case "medium":
		return "p2"
	case "low":
		return "p3"
	default:
		return "p3"
	}
}

var validTransitions = map[string][]string{
	"new":        {"in_review", "closed"},
	"in_review":  {"in_progress", "closed"},
	"in_progress": {"resolved"},
	"resolved":   {"closed", "reopened"},
	"closed":     {"reopened"},
	"reopened":   {"in_progress"},
}

func isValidTransition(from, to string) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func resetStoreForTests() {
	defaultStore = NewTicketStore()
	auditCounter = 0
}

var auditCounter int64

func newAuditEntryID() string {
	auditCounter++
	return fmt.Sprintf("audit-%d-%d", time.Now().UnixNano(), auditCounter)
}
