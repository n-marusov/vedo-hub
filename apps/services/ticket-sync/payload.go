// @ctx: Outbound payload conversion — VEDO ticket -> GitLab issue format
// @hlv:sec [INPUT_VALIDATION] — ticket data sanitized before conversion

package main

import (
	"fmt"
	"log/slog"
	"strings"
)

type GitLabIssue struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Labels      []string `json:"labels,omitempty"`
}

const gitlabDescriptionMaxLen = 1048576

func TicketToGitLabIssue(action SyncAction, data TicketData) GitLabIssue {
	slog.Info("payload.convert.enter",
		"ticket_id", data.ID,
		"action", string(action),
	)

	title := buildGitLabTitle(action, data)
	description := buildGitLabDescription(data)
	labels := buildGitLabLabels(action, data)

	// @hlv SYNC-PAYLOAD-TOO-LARGE — truncate description if over limit
	if len(description) > gitlabDescriptionMaxLen {
		description = description[:gitlabDescriptionMaxLen] + "\n\n[description truncated]"
		slog.Warn("payload.description_truncated",
			"ticket_id", data.ID,
			"max_len", gitlabDescriptionMaxLen,
		)
	}

	issue := GitLabIssue{
		Title:       title,
		Description: description,
		Labels:      labels,
	}

	slog.Info("payload.convert.exit",
		"ticket_id", data.ID,
		"title_len", len(title),
		"desc_len", len(description),
		"labels_count", len(labels),
	)
	return issue
}

func buildGitLabTitle(action SyncAction, data TicketData) string {
	prefix := ""
	switch action {
	case SyncActionCreate:
		prefix = "[VEDO] "
	case SyncActionUpdate:
		prefix = "[VEDO:UPD] "
	case SyncActionClose:
		prefix = "[VEDO:CLS] "
	case SyncActionReopen:
		prefix = "[VEDO:REO] "
	}
	return prefix + data.Title
}

func buildGitLabDescription(data TicketData) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("**VEDO Ticket**: %s\n", data.ID))
	b.WriteString(fmt.Sprintf("**Category**: %s\n", data.Category))
	b.WriteString(fmt.Sprintf("**Severity**: %s\n", data.UserSeverity))
	b.WriteString(fmt.Sprintf("**Priority**: %s\n", data.SystemPriority))
	b.WriteString(fmt.Sprintf("**Status**: %s\n", data.Status))
	b.WriteString(fmt.Sprintf("**Source**: %s\n", data.Source))
	b.WriteString(fmt.Sprintf("**Channel**: %s\n\n", data.Channel))
	b.WriteString("---\n\n")
	b.WriteString(data.Description)

	if data.StepsToReproduce != nil && *data.StepsToReproduce != "" {
		b.WriteString("\n\n### Steps to Reproduce\n\n")
		b.WriteString(*data.StepsToReproduce)
	}
	if data.ExpectedBehavior != nil && *data.ExpectedBehavior != "" {
		b.WriteString("\n\n### Expected Behavior\n\n")
		b.WriteString(*data.ExpectedBehavior)
	}

	return b.String()
}

func buildGitLabLabels(action SyncAction, data TicketData) []string {
	labels := make([]string, 0, 6)
	if data.Labels != nil {
		labels = append(labels, data.Labels...)
	}
	labels = append(labels, fmt.Sprintf("vedo:%s", data.Source))
	labels = append(labels, fmt.Sprintf("vedo:category:%s", data.Category))
	labels = append(labels, fmt.Sprintf("vedo:severity:%s", data.UserSeverity))
	labels = append(labels, fmt.Sprintf("vedo:priority:%s", data.SystemPriority))
	labels = append(labels, fmt.Sprintf("vedo:action:%s", string(action)))

	return labels
}
