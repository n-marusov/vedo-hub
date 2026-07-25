// @ctx: in-memory ticket store — CRUD, comments, deduplication
// @hlv:sec [INPUT_VALIDATION] — input validation before store operations

package ticketapi

import (
	"log/slog"
	"sync"
	"time"
)

type TicketStore struct {
	mu       sync.RWMutex
	tickets  map[string]*Ticket
	comments map[string][]TicketComment
}

func NewTicketStore() *TicketStore {
	return &TicketStore{
		tickets:  make(map[string]*Ticket),
		comments: make(map[string][]TicketComment),
	}
}

func (s *TicketStore) Create(req CreateTicketRequest, metadata TicketMetadata) (*Ticket, error) {
	if req.Title == "" || req.Description == "" {
		// @hlv TICKET-INVALID-REQUEST
		return nil, &AppError{Code: ErrInvalidRequest, Message: "Title and description are required", Ref: "T-02"}
	}
	if !IsValidCategory(req.Category) {
		return nil, &AppError{Code: ErrInvalidRequest, Message: "Invalid category: " + string(req.Category), Ref: "T-02"}
	}
	if !IsValidSeverity(req.UserSeverity) {
		return nil, &AppError{Code: ErrInvalidRequest, Message: "Invalid user_severity: " + string(req.UserSeverity), Ref: "T-02"}
	}
	if !IsValidSource(req.Source) {
		return nil, &AppError{Code: ErrInvalidRequest, Message: "Invalid source: " + string(req.Source), Ref: "T-02"}
	}
	if !IsValidChannel(req.Channel) {
		// @hlv TICKET-UNSUPPORTED-CHANNEL
		return nil, &AppError{Code: ErrUnsupportedChannel, Message: "Invalid channel: " + string(req.Channel), Ref: "T-47"}
	}
	if len(req.Attachments) > MaxAttachments {
		// @hlv TICKET-TOO-MANY-ATTACHMENTS
		return nil, &AppError{Code: ErrTooManyAttachments, Message: "Maximum of 10 attachments allowed", Ref: "T-03"}
	}
	for _, a := range req.Attachments {
		if a.Size > MaxAttachmentSize || a.Filename == "" {
			// @hlv TICKET-INVALID-ATTACHMENT
			return nil, &AppError{Code: ErrInvalidAttachment, Message: "Attachment size exceeds 10 MB or filename is invalid", Ref: "T-03"}
		}
	}

	now := time.Now().UTC()
	priority := DerivePriority(req.UserSeverity, req.Category)

	labels := make([]string, 0)
	if req.Source == TicketSourceTelemetry {
		// @hlv telemetry_label
		labels = append(labels, "source=telemetry")
	}

	ticket := &Ticket{
		ID:             NewUUID(),
		Source:         req.Source,
		Channel:        req.Channel,
		Status:         TicketStatusNew,
		Title:          req.Title,
		Description:    req.Description,
		Category:       req.Category,
		UserSeverity:   req.UserSeverity,
		SystemPriority: priority,
		Metadata:       metadata,
		Attachments:    req.Attachments,
		Labels:         labels,
		Comments:       make([]TicketComment, 0),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	s.mu.Lock()
	s.tickets[ticket.ID] = ticket
	// @hlv uuid_uniqueness
	s.mu.Unlock()

	slog.Info("ticket.created",
		"ticket_id", ticket.ID,
		"source", ticket.Source,
		"channel", ticket.Channel,
		"category", ticket.Category,
		"priority", ticket.SystemPriority,
	)

	return ticket, nil
}

func (s *TicketStore) GetByID(id string) (*Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, ok := s.tickets[id]
	if !ok {
		// @hlv TICKET-NOT-FOUND
		return nil, &AppError{Code: ErrNotFound, Message: "Ticket not found", Ref: "T-06"}
	}
	return ticket, nil
}

func (s *TicketStore) Update(id string, req UpdateTicketRequest) (*Ticket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, ok := s.tickets[id]
	if !ok {
		return nil, &AppError{Code: ErrNotFound, Message: "Ticket not found", Ref: "T-06"}
	}

	oldStatus := ticket.Status

	if req.Title != nil {
		ticket.Title = *req.Title
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.Category != nil {
		if !IsValidCategory(*req.Category) {
			return nil, &AppError{Code: ErrInvalidRequest, Message: "Invalid category", Ref: "T-02"}
		}
		ticket.Category = *req.Category
		ticket.SystemPriority = DerivePriority(ticket.UserSeverity, *req.Category)
	}
	if req.UserSeverity != nil {
		if !IsValidSeverity(*req.UserSeverity) {
			return nil, &AppError{Code: ErrInvalidRequest, Message: "Invalid severity", Ref: "T-02"}
		}
		ticket.UserSeverity = *req.UserSeverity
		ticket.SystemPriority = DerivePriority(*req.UserSeverity, ticket.Category)
	}
	if req.SystemPriority != nil {
		ticket.SystemPriority = *req.SystemPriority
	}
	if req.Status != nil {
		// @hlv status_transition_validity
		if !IsValidTransition(ticket.Status, *req.Status) {
			return nil, &AppError{Code: ErrConflict, Message: "Invalid status transition from " + string(ticket.Status) + " to " + string(*req.Status), Ref: "lifecycle status transitions (Figure 2)"}
		}
		ticket.Status = *req.Status
		if ticket.Status == TicketStatusClosed {
			now := time.Now().UTC()
			ticket.ClosedAt = &now
		}
	}
	if req.StepsToReproduce != nil {
		ticket.StepsToReproduce = req.StepsToReproduce
	}
	if req.ExpectedBehavior != nil {
		ticket.ExpectedBehavior = req.ExpectedBehavior
	}
	if req.ExpectedDuration != nil {
		ticket.ExpectedDuration = req.ExpectedDuration
	}
	if req.ActualDuration != nil {
		ticket.ActualDuration = req.ActualDuration
	}
	if req.Comment != nil {
		comment := TicketComment{
			Author:    "system",
			Text:      *req.Comment,
			CreatedAt: time.Now().UTC(),
		}
		if ticket.Comments == nil {
			ticket.Comments = make([]TicketComment, 0)
		}
		ticket.Comments = append(ticket.Comments, comment)
		ticket.UpdatedAt = time.Now().UTC()
	}

	ticket.UpdatedAt = time.Now().UTC()

	slog.Info("ticket.updated",
		"ticket_id", ticket.ID,
		"old_status", oldStatus,
		"new_status", ticket.Status,
		"channel", req.Channel,
	)
	// @hlv log_state_changes

	return ticket, nil
}

func (s *TicketStore) List(source *TicketSource, status *TicketStatus, category *TicketCategory, limit, offset int) ([]*Ticket, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*Ticket
	for _, t := range s.tickets {
		if source != nil && t.Source != *source {
			continue
		}
		if status != nil && t.Status != *status {
			continue
		}
		if category != nil && t.Category != *category {
			continue
		}
		filtered = append(filtered, t)
	}

	total := len(filtered)

	if offset > len(filtered) {
		return []*Ticket{}, total
	}
	filtered = filtered[offset:]

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}

	return filtered, total
}

func (s *TicketStore) AddComment(ticketID, author, text string) (*TicketComment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, ok := s.tickets[ticketID]
	if !ok {
		return nil, &AppError{Code: ErrNotFound, Message: "Ticket not found", Ref: "T-06"}
	}

	comment := TicketComment{
		Author:    author,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}
	ticket.Comments = append(ticket.Comments, comment)
	ticket.UpdatedAt = time.Now().UTC()

	slog.Info("ticket.comment_added",
		"ticket_id", ticketID,
		"author", author,
	)

	return &comment, nil
}

type AppError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Ref     string `json:"ticket_management_system_ref,omitempty"`
}

func (e *AppError) Error() string {
	return e.Code + ": " + e.Message
}
