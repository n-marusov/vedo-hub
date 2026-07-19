// @ctx: in-memory audit trail — append-only event log for ticket operations
// @hlv:sec [SECRET_HANDLING] — audit entries must not leak sensitive data in logs

package ticketapi

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type AuditStore struct {
	mu      sync.RWMutex
	entries []AuditEntry
}

func NewAuditStore() *AuditStore {
	return &AuditStore{
		entries: make([]AuditEntry, 0, 1024),
	}
}

func (a *AuditStore) Record(actor, role, command, ticketID, traceID, result string) AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := AuditEntry{
		ID:        NewUUID(),
		Actor:     actor,
		Role:      role,
		Command:   command,
		TicketID:  ticketID,
		Result:    result,
		Timestamp: time.Now().UTC(),
	}
	if traceID != "" {
		entry.TraceID = &traceID
	}

	a.entries = append(a.entries, entry)
	return entry
}

func (a *AuditStore) GetByTicketID(ticketID string) []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var result []AuditEntry
	for _, e := range a.entries {
		if e.TicketID == ticketID {
			result = append(result, e)
		}
	}
	return result
}

func (a *AuditStore) GetAll() []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]AuditEntry, len(a.entries))
	copy(result, a.entries)
	return result
}

func NewUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
