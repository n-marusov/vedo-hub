// @ctx: Sync engine — validate -> convert -> execute -> track -> respond
// @hlv Outbound payloads follow T-19 contract format
// @hlv Sync status tracked per ticket
// @hlv All attempts logged in audit trail
// @hlv VEDO Core is source of truth
// @hlv Sync transitions: pending -> in_progress -> {synced, failed, conflict}
// @hlv Error details preserved for retry

package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type SyncEngine struct {
	mu      sync.RWMutex
	records map[string]*SyncRecord
	client  GitLabClient
}

func NewSyncEngine(client GitLabClient) *SyncEngine {
	return &SyncEngine{
		records: make(map[string]*SyncRecord),
		client:  client,
	}
}

func (e *SyncEngine) Execute(req SyncRequest) SyncResponse {
	syncID := NewUUID()
	slog.Info("sync.execute.enter",
		"sync_id", syncID,
		"ticket_id", req.TicketID,
		"action", string(req.Action),
	)

	resp := SyncResponse{
		SyncID:     syncID,
		TicketID:   req.TicketID,
		SyncStatus: SyncStatusPending,
		Timestamp:  time.Now().UTC(),
	}

	// @hlv SYNC-INVALID-REQUEST — validate request
	if err := e.validateRequest(req); err != nil {
		slog.Error("sync.validation_failed",
			"sync_id", syncID,
			"ticket_id", req.TicketID,
			"error", err.Error(),
		)
		resp.SyncStatus = SyncStatusFailed
		errCode := ErrInvalidRequest
		resp.Error = &errCode
		e.recordAttempt(req.TicketID, req.GitLabProjectID, SyncStatusFailed, err.Error())
		slog.Info("sync.execute.exit",
			"sync_id", syncID,
			"status", string(resp.SyncStatus),
		)
		// @hlv log_all_errors
		return resp
	}

	// @hlv:sec [INPUT_VALIDATION] — ticket data validated before conversion
	resp.SyncStatus = SyncStatusInProgress
	e.recordAttempt(req.TicketID, req.GitLabProjectID, SyncStatusInProgress, "")

	issue := TicketToGitLabIssue(req.Action, req.TicketData)

	var result SyncResponse
	var syncErr error

	switch req.Action {
	case SyncActionCreate:
		result, syncErr = e.handleCreate(req, issue, syncID)
	case SyncActionUpdate:
		result, syncErr = e.handleUpdate(req, issue, syncID)
	case SyncActionClose:
		result, syncErr = e.handleClose(req, syncID)
	case SyncActionReopen:
		result, syncErr = e.handleReopen(req, syncID)
	default:
		result = resp
		errMsg := fmt.Sprintf("unsupported action: %s", req.Action)
		result.Error = &errMsg
		result.SyncStatus = SyncStatusFailed
		syncErr = fmt.Errorf("%s", errMsg)
	}

	if syncErr != nil {
		if result.SyncStatus != SyncStatusConflict {
			result.SyncStatus = SyncStatusFailed
		}
		if result.Error == nil || *result.Error == "" {
			errMsg := syncErr.Error()
			result.Error = &errMsg
		}
		result.Timestamp = time.Now().UTC()
		e.recordAttempt(req.TicketID, req.GitLabProjectID, SyncStatusFailed, syncErr.Error())
		slog.Error("sync.execute.failed",
			"sync_id", syncID,
			"ticket_id", req.TicketID,
			"error", syncErr.Error(),
		)
		// @hlv log_all_errors
	} else {
		e.recordAttempt(req.TicketID, req.GitLabProjectID, result.SyncStatus, "")
		slog.Info("sync.execute.success",
			"sync_id", syncID,
			"ticket_id", req.TicketID,
			"status", string(result.SyncStatus),
		)
	}

	slog.Info("sync.execute.exit",
		"sync_id", syncID,
		"status", string(result.SyncStatus),
	)
	return result
}

// @hlv:sec [INPUT_VALIDATION] — all user input validated before processing
func (e *SyncEngine) validateRequest(req SyncRequest) error {
	if req.TicketID == "" {
		return fmt.Errorf("ticket_id is required")
	}
	if !isValidUUID(req.TicketID) {
		return fmt.Errorf("ticket_id must be a valid UUID")
	}
	if !isValidAction(req.Action) {
		return fmt.Errorf("invalid action: %q, must be one of create/update/close/reopen", req.Action)
	}
	if req.TicketData.Title == "" {
		return fmt.Errorf("ticket_data.title is required")
	}
	if req.GitLabProjectID == "" {
		return fmt.Errorf("gitlab_project_id is required")
	}
	return nil
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	return true
}

func isValidAction(a SyncAction) bool {
	for _, va := range ValidActions {
		if a == va {
			return true
		}
	}
	return false
}

func (e *SyncEngine) handleCreate(req SyncRequest, issue GitLabIssue, syncID string) (SyncResponse, error) {
	slog.Info("sync.create.enter",
		"sync_id", syncID,
		"ticket_id", req.TicketID,
	)

	record := e.getRecord(req.TicketID)
	if record != nil && record.GitLabIssueID != "" {
		slog.Warn("sync.create.already_exists",
			"sync_id", syncID,
			"ticket_id", req.TicketID,
			"existing_issue_id", record.GitLabIssueID,
		)
		errCode := ErrConflict
		return e.buildSyncResponse(req, &record.GitLabIssueID, &record.GitLabIssueIID,
			SyncStatusConflict, &errCode, syncID), nil
		// @hlv SYNC-CONFLICT
	}

	result, err := e.client.CreateIssue(req.GitLabProjectID, issue)
	if err != nil {
		return e.detectErrorResponse(err, req, syncID)
	}

	e.storeRecord(req.TicketID, result.IssueID, result.IssueIID, req.GitLabProjectID)

	return e.buildSyncResponse(req, &result.IssueID, &result.IssueIID,
		SyncStatusSynced, nil, syncID), nil
}

func (e *SyncEngine) handleUpdate(req SyncRequest, issue GitLabIssue, syncID string) (SyncResponse, error) {
	slog.Info("sync.update.enter",
		"sync_id", syncID,
		"ticket_id", req.TicketID,
	)

	record := e.getRecord(req.TicketID)
	if record == nil || record.GitLabIssueID == "" {
		errCode := ErrInvalidRequest
		return e.buildSyncResponse(req, nil, nil,
			SyncStatusFailed, &errCode, syncID), fmt.Errorf("no gitlab issue found for ticket")
	}

	if record.Status == SyncStatusConflict {
		errCode := ErrConflict
		return e.buildSyncResponse(req, &record.GitLabIssueID, &record.GitLabIssueIID,
			SyncStatusConflict, &errCode, syncID), fmt.Errorf("conflict detected")
		// @hlv SYNC-CONFLICT
	}

	result, err := e.client.UpdateIssue(req.GitLabProjectID, record.GitLabIssueID, issue)
	if err != nil {
		return e.detectErrorResponse(err, req, syncID)
	}

	e.storeRecord(req.TicketID, result.IssueID, result.IssueIID, req.GitLabProjectID)

	return e.buildSyncResponse(req, &result.IssueID, &result.IssueIID,
		SyncStatusSynced, nil, syncID), nil
}

func (e *SyncEngine) handleClose(req SyncRequest, syncID string) (SyncResponse, error) {
	slog.Info("sync.close.enter",
		"sync_id", syncID,
		"ticket_id", req.TicketID,
	)

	record := e.getRecord(req.TicketID)
	if record == nil || record.GitLabIssueID == "" {
		errCode := ErrInvalidRequest
		return e.buildSyncResponse(req, nil, nil,
			SyncStatusFailed, &errCode, syncID), fmt.Errorf("no gitlab issue found for ticket")
	}

	if record.Status == SyncStatusConflict {
		errCode := ErrConflict
		return e.buildSyncResponse(req, &record.GitLabIssueID, &record.GitLabIssueIID,
			SyncStatusConflict, &errCode, syncID), fmt.Errorf("conflict detected")
		// @hlv SYNC-CONFLICT
	}

	err := e.client.CloseIssue(req.GitLabProjectID, record.GitLabIssueID)
	if err != nil {
		return e.detectErrorResponse(err, req, syncID)
	}

	e.storeRecord(req.TicketID, record.GitLabIssueID, record.GitLabIssueIID, req.GitLabProjectID)

	return e.buildSyncResponse(req, &record.GitLabIssueID, &record.GitLabIssueIID,
		SyncStatusSynced, nil, syncID), nil
}

func (e *SyncEngine) handleReopen(req SyncRequest, syncID string) (SyncResponse, error) {
	slog.Info("sync.reopen.enter",
		"sync_id", syncID,
		"ticket_id", req.TicketID,
	)

	record := e.getRecord(req.TicketID)
	if record == nil || record.GitLabIssueID == "" {
		errCode := ErrInvalidRequest
		return e.buildSyncResponse(req, nil, nil,
			SyncStatusFailed, &errCode, syncID), fmt.Errorf("no gitlab issue found for ticket")
	}

	err := e.client.ReopenIssue(req.GitLabProjectID, record.GitLabIssueID)
	if err != nil {
		return e.detectErrorResponse(err, req, syncID)
	}

	e.storeRecord(req.TicketID, record.GitLabIssueID, record.GitLabIssueIID, req.GitLabProjectID)

	return e.buildSyncResponse(req, &record.GitLabIssueID, &record.GitLabIssueIID,
		SyncStatusSynced, nil, syncID), nil
}

func (e *SyncEngine) detectErrorResponse(err error, req SyncRequest, syncID string) (SyncResponse, error) {
	errMsg := err.Error()

	var status SyncStatus
	var httpErr string

	switch {
	case contains(errMsg, "unavailable"):
		status = SyncStatusFailed
		httpErr = ErrGitLabUnavailable
	case contains(errMsg, "rate limited"):
		status = SyncStatusFailed
		httpErr = ErrRateLimited
	case contains(errMsg, "conflict"):
		status = SyncStatusConflict
		httpErr = ErrConflict
	default:
		status = SyncStatusFailed
		httpErr = ErrGitLabUnavailable
	}

	return e.buildSyncResponse(req, nil, nil, status, &httpErr, syncID), fmt.Errorf("%s", errMsg)
}

func (e *SyncEngine) buildSyncResponse(req SyncRequest, issueID, issueIID *string, status SyncStatus, errMsg *string, syncID string) SyncResponse {
	var url *string
	if issueID != nil && issueIID != nil {
		u := fmt.Sprintf("https://gitlab.example.com/%s/-/issues/%s", req.GitLabProjectID, *issueIID)
		url = &u
	}

	return SyncResponse{
		SyncID:         syncID,
		TicketID:       req.TicketID,
		GitLabIssueID:  issueID,
		GitLabIssueIID: issueIID,
		GitLabIssueURL: url,
		SyncStatus:     status,
		Error:          errMsg,
		Timestamp:      time.Now().UTC(),
	}
}

func (e *SyncEngine) getRecord(ticketID string) *SyncRecord {
	e.mu.RLock()
	defer e.mu.RUnlock()
	rec, ok := e.records[ticketID]
	if !ok {
		return nil
	}
	return rec
}

func (e *SyncEngine) storeRecord(ticketID, issueID, issueIID, projectID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	if rec, ok := e.records[ticketID]; ok {
		rec.GitLabIssueID = issueID
		rec.GitLabIssueIID = issueIID
		rec.Status = SyncStatusSynced
		rec.UpdatedAt = now
		return
	}

	e.records[ticketID] = &SyncRecord{
		TicketID:        ticketID,
		GitLabIssueID:   issueID,
		GitLabIssueIID:  issueIID,
		GitLabProjectID: projectID,
		Status:          SyncStatusSynced,
		Attempts:        0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// @hlv All attempts logged in audit trail
// @hlv Error details preserved for retry
func (e *SyncEngine) recordAttempt(ticketID, projectID string, status SyncStatus, errDetail string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	if rec, ok := e.records[ticketID]; ok {
		// @hlv Sync transitions: pending -> in_progress -> {synced, failed, conflict}
		// Preserve sticky terminal statuses (conflict) from being overwritten
		if status == SyncStatusInProgress && rec.Status == SyncStatusConflict {
			rec.Attempts++
			return
		}
		rec.Status = status
		rec.Attempts++
		rec.ErrorDetail = errDetail
		rec.UpdatedAt = now
		return
	}

	e.records[ticketID] = &SyncRecord{
		TicketID:        ticketID,
		GitLabProjectID: projectID,
		Status:          status,
		ErrorDetail:     errDetail,
		Attempts:        1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	// @hlv log_state_mutations
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
