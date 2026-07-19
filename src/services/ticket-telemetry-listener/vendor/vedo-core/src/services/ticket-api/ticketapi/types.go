// @hlv:artifact ticket-api-code implements TICKET-CORE-001
// @ctx: domain types and error codes for ticket API — maps TICKET-CORE-001 contract

package ticketapi

import "time"

type TicketSource string
type TicketChannel string
type TicketStatus string
type TicketCategory string
type TicketUserSeverity string
type TicketSystemPriority string

const (
	TicketSourceManual    TicketSource = "manual"
	TicketSourceTelemetry TicketSource = "telemetry"

	TicketChannelUI        TicketChannel = "ui"
	TicketChannelCLI       TicketChannel = "cli"
	TicketChannelTelemetry TicketChannel = "telemetry"

	TicketStatusNew        TicketStatus = "new"
	TicketStatusInReview   TicketStatus = "in_review"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
	TicketStatusReopened   TicketStatus = "reopened"

	TicketCategoryBug           TicketCategory = "bug"
	TicketCategoryPerformance   TicketCategory = "performance"
	TicketCategoryQuestion      TicketCategory = "question"
	TicketCategoryFeature       TicketCategory = "feature"
	TicketCategoryDocumentation TicketCategory = "documentation"
	TicketCategoryAccess        TicketCategory = "access"
	TicketCategoryNeedsAnalysis TicketCategory = "needs_analysis"

	TicketSeverityCritical TicketUserSeverity = "critical"
	TicketSeverityHigh     TicketUserSeverity = "high"
	TicketSeverityMedium   TicketUserSeverity = "medium"
	TicketSeverityLow      TicketUserSeverity = "low"

	TicketPriorityP0 TicketSystemPriority = "p0"
	TicketPriorityP1 TicketSystemPriority = "p1"
	TicketPriorityP2 TicketSystemPriority = "p2"
	TicketPriorityP3 TicketSystemPriority = "p3"
)

var ValidStatusTransitions = map[TicketStatus][]TicketStatus{
	TicketStatusNew:        {TicketStatusInReview, TicketStatusClosed},
	TicketStatusInReview:   {TicketStatusInProgress, TicketStatusClosed},
	TicketStatusInProgress: {TicketStatusResolved},
	TicketStatusResolved:   {TicketStatusClosed, TicketStatusReopened},
	TicketStatusClosed:     {TicketStatusReopened},
	TicketStatusReopened:   {TicketStatusInProgress},
}

type TicketAttachment struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
}

type TicketComment struct {
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type TicketMetadata struct {
	VedoVersion string  `json:"vedo_version"`
	Environment string  `json:"environment"`
	UserID      string  `json:"user_id"`
	TraceID     *string `json:"trace_id"`
	PageURL     *string `json:"page_url"`
	UserAgent   string  `json:"user_agent"`
}

type Ticket struct {
	ID               string               `json:"id"`
	ExternalID       *string              `json:"external_id"`
	Source           TicketSource         `json:"source"`
	Channel          TicketChannel        `json:"channel"`
	Status           TicketStatus         `json:"status"`
	Title            string               `json:"title"`
	Description      string               `json:"description"`
	Category         TicketCategory       `json:"category"`
	UserSeverity     TicketUserSeverity   `json:"user_severity"`
	SystemPriority   TicketSystemPriority `json:"system_priority"`
	Metadata         TicketMetadata       `json:"metadata"`
	StepsToReproduce *string              `json:"steps_to_reproduce"`
	ExpectedBehavior *string              `json:"expected_behavior"`
	ExpectedDuration *string              `json:"expected_duration"`
	ActualDuration   *string              `json:"actual_duration"`
	Attachments      []TicketAttachment   `json:"attachments"`
	TelemetryLink    *string              `json:"telemetry_link"`
	SourceTicketID   *string              `json:"source_ticket_id"`
	DuplicateOf      *string              `json:"duplicate_of"`
	Labels           []string             `json:"labels"`
	Comments         []TicketComment      `json:"comments"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	ClosedAt         *time.Time           `json:"closed_at"`
}

type CreateTicketRequest struct {
	Title        string             `json:"title" binding:"required"`
	Description  string             `json:"description" binding:"required"`
	Category     TicketCategory     `json:"category" binding:"required"`
	UserSeverity TicketUserSeverity `json:"user_severity" binding:"required"`
	Attachments  []TicketAttachment `json:"attachments,omitempty"`
	Source       TicketSource       `json:"source" binding:"required"`
	Channel      TicketChannel      `json:"channel" binding:"required"`
}

type UpdateTicketRequest struct {
	Status           *TicketStatus         `json:"status,omitempty"`
	Title            *string               `json:"title,omitempty"`
	Description      *string               `json:"description,omitempty"`
	Category         *TicketCategory       `json:"category,omitempty"`
	UserSeverity     *TicketUserSeverity   `json:"user_severity,omitempty"`
	SystemPriority   *TicketSystemPriority `json:"priority,omitempty"`
	StepsToReproduce *string               `json:"steps_to_reproduce,omitempty"`
	ExpectedBehavior *string               `json:"expected_behavior,omitempty"`
	ExpectedDuration *string               `json:"expected_duration,omitempty"`
	ActualDuration   *string               `json:"actual_duration,omitempty"`
	Channel          TicketChannel         `json:"channel,omitempty"`
	Comment          *string               `json:"comment,omitempty"`
}

type AuditEntry struct {
	ID        string    `json:"id"`
	Actor     string    `json:"actor"`
	Role      string    `json:"role"`
	Command   string    `json:"command"`
	TicketID  string    `json:"ticket_id"`
	TraceID   *string   `json:"trace_id"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Ref     string `json:"ticket_management_system_ref,omitempty"`
}

const (
	ErrInvalidRequest     = "TICKET-INVALID-REQUEST"
	ErrUnauthorized       = "TICKET-UNAUTHORIZED"
	ErrForbidden          = "TICKET-FORBIDDEN"
	ErrNotFound           = "TICKET-NOT-FOUND"
	ErrConflict           = "TICKET-CONFLICT"
	ErrTooManyAttachments = "TICKET-TOO-MANY-ATTACHMENTS"
	ErrInvalidAttachment  = "TICKET-INVALID-ATTACHMENT"
	ErrUnsupportedChannel = "TICKET-UNSUPPORTED-CHANNEL"
)

const MaxAttachments = 10
const MaxAttachmentSize int64 = 10 * 1024 * 1024
const DedupeWindow = 24 * time.Hour
