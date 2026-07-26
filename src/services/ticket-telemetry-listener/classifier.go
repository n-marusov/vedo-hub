// @ctx: telemetry event classifier — Python client stub with deterministic Go fallback
// @hlv:sec [INPUT_VALIDATION] — classification determines ticket category

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"vedo-core/src/services/ticket-api/ticketapi"
)

type ClassifyResult struct {
	Category ticketapi.TicketCategory
	Source   string
}

type PythonClassifierClient struct {
	URL    string
	Client *http.Client
}

type GoFallbackClassifier struct{}

func NewPythonClassifierClient(url string) *PythonClassifierClient {
	return &PythonClassifierClient{
		URL:    url,
		Client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *PythonClassifierClient) Classify(_ TelemetryEvent) (*ClassifyResult, error) {
	url := fmt.Sprintf("%s/classify", strings.TrimRight(p.URL, "/"))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("python classifier unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("python classifier returned status %d", resp.StatusCode)
	}
	return nil, fmt.Errorf("python classifier stub: not implemented")
}

func NewGoFallbackClassifier() *GoFallbackClassifier {
	return &GoFallbackClassifier{}
}

func (g *GoFallbackClassifier) Classify(event TelemetryEvent) *ClassifyResult {
	category := g.determineCategory(event)
	slog.Info("classifier.go_fallback",
		"event_type", event.EventType,
		"source", event.Source,
		"category", category,
	)
	return &ClassifyResult{
		Category: category,
		Source:   "go_fallback",
	}
}

func (g *GoFallbackClassifier) determineCategory(event TelemetryEvent) ticketapi.TicketCategory {
	switch event.EventType {
	case "alert":
		if isSourceRelated(event.Source, []string{"sparql", "api", "gateway", "endpoint", "service"}) {
			return ticketapi.TicketCategoryPerformance
		}
		if isSourceRelated(event.Source, []string{"redis", "cache", "memcache", "cluster", "neo4j"}) {
			return ticketapi.TicketCategoryAccess
		}
		return ticketapi.TicketCategoryNeedsAnalysis
	case "log":
		return ticketapi.TicketCategoryBug
	case "metric":
		return ticketapi.TicketCategoryNeedsAnalysis
	default:
		return ticketapi.TicketCategoryNeedsAnalysis
	}
}

func isSourceRelated(source string, keywords []string) bool {
	sourceLower := strings.ToLower(source)
	for _, kw := range keywords {
		if strings.Contains(sourceLower, kw) {
			return true
		}
	}
	return false
}
