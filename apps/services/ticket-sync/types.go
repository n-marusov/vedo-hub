// @hlv:artifact ticket-sync-code implements TICKET-SYNC-001
// @ctx: domain types, error codes, and sync status types for ticket sync service
// @hlv Outbound payloads follow T-19 contract format
// @hlv Sync status tracked per ticket
// @hlv All attempts logged in audit trail
// @hlv VEDO Core is source of truth
// @hlv Sync transitions: pending -> in_progress -> {synced, failed, conflict}
// @hlv Error details preserved for retry

package main

import "time"

type SyncStatus string

const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusSynced     SyncStatus = "synced"
	SyncStatusFailed     SyncStatus = "failed"
	SyncStatusConflict   SyncStatus = "conflict"
)

var ValidSyncTransitions = map[SyncStatus][]SyncStatus{
	SyncStatusPending:    {SyncStatusInProgress},
	SyncStatusInProgress: {SyncStatusSynced, SyncStatusFailed, SyncStatusConflict},
}

type SyncAction string

const (
	SyncActionCreate SyncAction = "create"
	SyncActionUpdate SyncAction = "update"
	SyncActionClose  SyncAction = "close"
	SyncActionReopen SyncAction = "reopen"
)

var ValidActions = []SyncAction{SyncActionCreate, SyncActionUpdate, SyncActionClose, SyncActionReopen}

type TicketData struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Category         string   `json:"category"`
	UserSeverity     string   `json:"user_severity"`
	SystemPriority   string   `json:"system_priority"`
	Source           string   `json:"source"`
	Channel          string   `json:"channel"`
	Status           string   `json:"status"`
	Labels           []string `json:"labels,omitempty"`
	ExternalID       *string  `json:"external_id,omitempty"`
	StepsToReproduce *string  `json:"steps_to_reproduce,omitempty"`
	ExpectedBehavior *string  `json:"expected_behavior,omitempty"`
}

type SyncRequest struct {
	TicketID        string     `json:"ticket_id" binding:"required"`
	Action          SyncAction `json:"action" binding:"required"`
	TicketData      TicketData `json:"ticket_data" binding:"required"`
	GitLabProjectID string     `json:"gitlab_project_id" binding:"required"`
}

type SyncResponse struct {
	SyncID         string     `json:"sync_id"`
	TicketID       string     `json:"ticket_id"`
	GitLabIssueID  *string    `json:"gitlab_issue_id"`
	GitLabIssueIID *string    `json:"gitlab_issue_iid"`
	GitLabIssueURL *string    `json:"gitlab_issue_url"`
	SyncStatus     SyncStatus `json:"sync_status"`
	Error          *string    `json:"error"`
	Timestamp      time.Time  `json:"timestamp"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Ref     string `json:"ticket_management_system_ref,omitempty"`
}

type SyncRecord struct {
	TicketID        string
	GitLabIssueID   string
	GitLabIssueIID  string
	GitLabProjectID string
	Status          SyncStatus
	ErrorDetail     string
	Attempts        int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

const (
	ErrInvalidRequest    = "SYNC-INVALID-REQUEST"
	ErrUnauthorized      = "SYNC-UNAUTHORIZED"
	ErrGitLabUnavailable = "SYNC-GITLAB-UNAVAILABLE"
	ErrRateLimited       = "SYNC-RATE-LIMITED"
	ErrConflict          = "SYNC-CONFLICT"
	ErrPayloadTooLarge   = "SYNC-PAYLOAD-TOO-LARGE"
)
