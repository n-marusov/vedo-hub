// Package main provides notification wiring for the commenting service.
// Supports in-process SSE streaming for real-time comment delivery and
// structured event logging for downstream integration (ticket-notifier, etc.).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// CommentEventType represents the type of comment event.
type CommentEventType string

const (
	EventCommentCreated CommentEventType = "comment.created"
	EventCommentUpdated CommentEventType = "comment.updated"
	EventCommentDeleted CommentEventType = "comment.deleted"
)

// CommentEvent represents a comment lifecycle event for subscribers.
type CommentEvent struct {
	Type       CommentEventType `json:"type"`
	CommentID  string           `json:"comment_id"`
	OntologyID string           `json:"ontology_id"`
	EntityID   string           `json:"entity_id"`
	AuthorID   string           `json:"author_id"`
	Timestamp  time.Time        `json:"timestamp"`
	Payload    *Comment         `json:"payload,omitempty"`
}

// EventBus manages comment event subscribers and broadcasting.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string]chan CommentEvent
	nextID      int
}

// newEventBus creates a new EventBus instance.
func newEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]chan CommentEvent),
	}
}

// Subscribe registers a new subscriber channel for comment events.
// Returns a unique subscriber ID and the channel for receiving events.
func (eb *EventBus) Subscribe(bufferSize int) (string, <-chan CommentEvent) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	id := fmt.Sprintf("sub_%d_%d", time.Now().UnixNano(), eb.nextID)
	eb.nextID++
	ch := make(chan CommentEvent, bufferSize)
	eb.subscribers[id] = ch

	slog.Debug("Event subscriber registered", "subscriber_id", id, "buffer_size", bufferSize)
	return id, ch
}

// Unsubscribe removes a subscriber by ID and closes its channel.
func (eb *EventBus) Unsubscribe(id string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if ch, ok := eb.subscribers[id]; ok {
		close(ch)
		delete(eb.subscribers, id)
		slog.Debug("Event subscriber unregistered", "subscriber_id", id)
	}
}

// Publish broadcasts a CommentEvent to all active subscribers.
// Slow subscribers are dropped after a short timeout to avoid blocking.
func (eb *EventBus) Publish(event CommentEvent) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	slog.Debug("Publishing comment event",
		"event_type", event.Type,
		"comment_id", event.CommentID,
		"ontology_id", event.OntologyID,
		"subscriber_count", len(eb.subscribers),
	)

	for id, ch := range eb.subscribers {
		select {
		case ch <- event:
			// Event delivered
		case <-time.After(100 * time.Millisecond):
			// Slow subscriber — drop event to avoid blocking
			slog.Warn("Dropping event for slow subscriber",
				"subscriber_id", id,
				"event_type", event.Type,
			)
		}
	}

	// Also log as structured event for downstream integration (RabbitMQ, etc.)
	emitStructuredEvent(event)
}

// emitStructuredEvent logs a structured JSON event for downstream consumers
// (e.g., ticket-notifier, log aggregator). In future iterations, this will
// publish to a RabbitMQ exchange.
func emitStructuredEvent(event CommentEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal event for structured logging", "error", err)
		return
	}

	slog.Info("Comment event",
		"event", string(payload),
		"event_type", event.Type,
	)
}

// emitCommentCreated publishes a created event and logs it.
func emitCommentCreated(bus *EventBus, comment *Comment) {
	if bus == nil {
		return
	}
	bus.Publish(CommentEvent{
		Type:       EventCommentCreated,
		CommentID:  comment.ID,
		OntologyID: comment.OntologyID,
		EntityID:   comment.EntityID,
		AuthorID:   comment.AuthorID,
		Timestamp:  time.Now().UTC(),
		Payload:    comment,
	})
}

// emitCommentUpdated publishes an updated event.
func emitCommentUpdated(bus *EventBus, comment *Comment) {
	if bus == nil {
		return
	}
	bus.Publish(CommentEvent{
		Type:       EventCommentUpdated,
		CommentID:  comment.ID,
		OntologyID: comment.OntologyID,
		EntityID:   comment.EntityID,
		AuthorID:   comment.AuthorID,
		Timestamp:  time.Now().UTC(),
		Payload:    comment,
	})
}

// emitCommentDeleted publishes a deleted event.
func emitCommentDeleted(bus *EventBus, commentID, ontologyID string) {
	if bus == nil {
		return
	}
	bus.Publish(CommentEvent{
		Type:       EventCommentDeleted,
		CommentID:  commentID,
		OntologyID: ontologyID,
		Timestamp:  time.Now().UTC(),
	})
}

// SSEBroker manages Server-Sent Events connections for live comment streaming.
type SSEBroker struct {
	bus *EventBus
}

// newSSEBroker creates an SSEBroker that subscribes to the given EventBus.
func newSSEBroker(bus *EventBus) *SSEBroker {
	return &SSEBroker{bus: bus}
}

// ServeSSE handles an HTTP connection upgrade to Server-Sent Events.
// It subscribes to the EventBus and streams events as SSE format.
// The connection is closed when the request context is cancelled.
func (b *SSEBroker) ServeSSE(ctx context.Context, ontologyID string, eventCh chan<- SSEEvent, errCh chan<- error) {
	subID, ch := b.bus.Subscribe(100)
	defer b.bus.Unsubscribe(subID)

	slog.Info("SSE client connected", "subscriber_id", subID, "ontology_id", ontologyID)

	for {
		select {
		case <-ctx.Done():
			slog.Info("SSE client disconnected", "subscriber_id", subID)
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			// Only emit events for the requested ontology (or all if ontologyID is empty)
			if ontologyID != "" && event.OntologyID != ontologyID {
				continue
			}

			data, err := json.Marshal(event)
			if err != nil {
				slog.Error("Failed to marshal SSE event", "error", err)
				continue
			}

			select {
			case eventCh <- SSEEvent{
				Event: string(event.Type),
				Data:  string(data),
			}:
			case <-ctx.Done():
				return
			}
		}
	}
}

// SSEEvent represents a single Server-Sent Event message.
type SSEEvent struct {
	Event string
	Data  string
}
