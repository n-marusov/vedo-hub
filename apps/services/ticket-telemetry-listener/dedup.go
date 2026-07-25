// @ctx: in-memory deduplication engine — 24h window by dedupe_signature
// @hlv:sec [INPUT_VALIDATION] — dedupe check before ticket creation

package main

import (
	"log/slog"
	"sync"
	"time"

	"vedo-core/src/services/ticket-api/ticketapi"
)

type DedupStore struct {
	mu      sync.RWMutex
	entries map[string]*DedupEntry
}

func NewDedupStore() *DedupStore {
	return &DedupStore{
		entries: make(map[string]*DedupEntry),
	}
}

func (d *DedupStore) Check(signature string, timestamp time.Time) *string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entry, ok := d.entries[signature]
	if !ok {
		return nil
	}

	if timestamp.Sub(entry.CreatedAt) <= ticketapi.DedupeWindow {
		return &entry.TicketID
	}
	return nil
}

func (d *DedupStore) Record(signature, ticketID string, createdAt time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.entries[signature] = &DedupEntry{
		TicketID:  ticketID,
		CreatedAt: createdAt,
	}
	slog.Info("dedup.recorded",
		"signature", signature,
		"ticket_id", ticketID,
		"created_at", createdAt,
	)
}

func (d *DedupStore) GetEntry(signature string) *DedupEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.entries[signature]
}

func (d *DedupStore) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.entries)
}
