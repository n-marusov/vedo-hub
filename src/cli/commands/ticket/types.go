package ticket

import "time"

// @hlv CLI_INVALID_COMMAND
const ErrInvalidCommand = "CLI-INVALID-COMMAND"

// @hlv CLI_MISSING_REQUIRED_FIELD
const ErrMissingRequiredField = "CLI-MISSING-REQUIRED-FIELD"

// @hlv CLI_UNAUTHORIZED
const ErrUnauthorized = "CLI-UNAUTHORIZED"

// @hlv CLI_FORBIDDEN
const ErrForbidden = "CLI-FORBIDDEN"

// @hlv CLI_TICKET_NOT_FOUND
const ErrTicketNotFound = "CLI-TICKET-NOT-FOUND"

// @hlv CLI_INVALID_STATE_TRANSITION
const ErrInvalidStateTransition = "CLI-INVALID-STATE-TRANSITION"

// @hlv CLI_DELETE_GUARDRAIL_VIOLATION
const ErrDeleteGuardrailViolation = "CLI-DELETE-GUARDRAIL-VIOLATION"

// @hlv CLI_AUDIT_LOG_FAILURE
const ErrAuditLogFailure = "CLI-AUDIT-LOG-FAILURE"

type TicketCliInput struct {
	Command      string `json:"command"`
	TicketID     string `json:"ticket_id,omitempty"`
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	Category     string `json:"category,omitempty"`
	Severity     string `json:"user_severity,omitempty"`
	Priority     string `json:"priority,omitempty"`
	Assignee     string `json:"assignee,omitempty"`
	Comment      string `json:"comment,omitempty"`
	Resolution   string `json:"resolution,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Actor        string `json:"actor,omitempty"`
	Role         string `json:"role,omitempty"`
	EnvGuard     bool   `json:"env_guard,omitempty"`
	TypedConfirm bool   `json:"typed_confirmation,omitempty"`
	MFAConfirmed bool   `json:"mfa_confirmed,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
	Channel      string `json:"channel,omitempty"`
	Source       string `json:"source,omitempty"`
}

type TicketCliOutput struct {
	Success    bool         `json:"success"`
	Ticket     *TicketData  `json:"ticket,omitempty"`
	Tickets    []TicketData `json:"tickets,omitempty"`
	Message    string       `json:"message"`
	AuditEntry *AuditEntry  `json:"audit_entry,omitempty"`
	Error      *CliError    `json:"error,omitempty"`
}

type TicketMetadata struct {
	VedoVersion string  `json:"vedo_version"`
	Environment string  `json:"environment"`
	UserID      string  `json:"user_id"`
	TraceID     *string `json:"trace_id"`
	PageURL     *string `json:"page_url"`
	UserAgent   string  `json:"user_agent"`
}

type TicketCommentData struct {
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type TicketData struct {
	ID               string              `json:"id"`
	Source           string              `json:"source"`
	Channel          string              `json:"channel"`
	Status           string              `json:"status"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Category         string              `json:"category"`
	UserSeverity     string              `json:"user_severity"`
	SystemPriority   string              `json:"system_priority"`
	Metadata         TicketMetadata      `json:"metadata"`
	StepsToReproduce *string             `json:"steps_to_reproduce"`
	ExpectedBehavior *string             `json:"expected_behavior"`
	ExpectedDuration *string             `json:"expected_duration"`
	ActualDuration   *string             `json:"actual_duration"`
	Attachments      []interface{}       `json:"attachments"`
	TelemetryLink    *string             `json:"telemetry_link"`
	SourceTicketID   *string             `json:"source_ticket_id"`
	DuplicateOf      *string             `json:"duplicate_of"`
	Labels           []string            `json:"labels"`
	Comments         []TicketCommentData `json:"comments"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	ClosedAt         *time.Time          `json:"closed_at"`
}

type AuditEntry struct {
	ID        string    `json:"id"`
	Actor     string    `json:"actor"`
	Role      string    `json:"role"`
	Command   string    `json:"command"`
	TicketID  string    `json:"ticket_id"`
	TraceID   string    `json:"trace_id"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

type CliError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
