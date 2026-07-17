package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ai_orchestrationv1 "vedo-core/src/services/shared/proto/ai-orchestration/v1"
)

// ontologyTemplate represents a loaded ontology template from a JSON file.
type ontologyTemplate struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Domain      string               `json:"domain"`
	Steps       []modelsSequenceStep `json:"steps"`
}

// templateStore manages loaded ontology templates.
type templateStore struct {
	mu        sync.RWMutex
	templates map[string]*ontologyTemplate
}

var globalTemplateStore = newTemplateStore()

func newTemplateStore() *templateStore {
	store := &templateStore{
		templates: make(map[string]*ontologyTemplate),
	}
	if err := store.loadTemplates(); err != nil {
		slog.Warn("handler.templates.load.error", "error", err)
	}
	return store
}

// templateDir returns the path to the templates directory.
func templateDir() string {
	// Relative to the service binary or CWD
	candidates := []string{
		"internal/templates/ontologies",
		"../internal/templates/ontologies",
		"src/services/ai-orchestration-service/internal/templates/ontologies",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return "internal/templates/ontologies"
}

func (s *templateStore) loadTemplates() error {
	dir := templateDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading template directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Warn("handler.templates.load.read_error", "file", entry.Name(), "error", err)
			continue
		}

		var tpl ontologyTemplate
		if err := json.Unmarshal(data, &tpl); err != nil {
			slog.Warn("handler.templates.load.parse_error", "file", entry.Name(), "error", err)
			continue
		}

		if tpl.ID == "" {
			slog.Warn("handler.templates.load.missing_id", "file", entry.Name())
			continue
		}

		s.mu.Lock()
		s.templates[tpl.ID] = &tpl
		s.mu.Unlock()

		slog.Debug("handler.templates.load.success",
			"id", tpl.ID,
			"name", tpl.Name,
			"steps", len(tpl.Steps),
		)
	}

	slog.Info("handler.templates.load.complete",
		"count", len(s.templates),
		"dir", dir,
	)
	return nil
}

// ListTemplates returns available ontology templates, optionally filtered by domain.
func ListTemplates(ctx context.Context, req *ai_orchestrationv1.ListTemplatesRequest) (*ai_orchestrationv1.ListTemplatesResponse, error) {
	start := time.Now()

	store := globalTemplateStore
	store.mu.RLock()
	defer store.mu.RUnlock()

	protoTemplates := make([]*ai_orchestrationv1.OntologyTemplate, 0, len(store.templates))
	for _, tpl := range store.templates {
		if req.GetDomainFilter() != "" && tpl.Domain != req.GetDomainFilter() {
			continue
		}

		classCount := 0
		propCount := 0
		indivCount := 0
		for _, step := range tpl.Steps {
			switch strings.ToUpper(step.Operation) {
			case "CREATE_CLASS":
				classCount++
			case "CREATE_OBJECT_PROPERTY", "CREATE_DATATYPE_PROPERTY":
				propCount++
			case "CREATE_INDIVIDUAL":
				indivCount++
			}
		}

		// Convert template steps to proto sequence steps
		protoSteps := make([]*ai_orchestrationv1.SequenceStep, len(tpl.Steps))
		for i, step := range tpl.Steps {
			protoSteps[i] = toProtoSequenceStep(step)
		}

		protoTemplates = append(protoTemplates, &ai_orchestrationv1.OntologyTemplate{
			Id:              tpl.ID,
			Name:            tpl.Name,
			Description:     tpl.Description,
			Domain:          tpl.Domain,
			Steps:           protoSteps,
			ClassCount:      int32(classCount),
			PropertyCount:   int32(propCount),
			IndividualCount: int32(indivCount),
		})
	}

	slog.Debug("handler.templates.list",
		"count", len(protoTemplates),
		"domain_filter", req.GetDomainFilter(),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &ai_orchestrationv1.ListTemplatesResponse{
		Templates: protoTemplates,
	}, nil
}

// GetTemplate returns a single template by ID.
func GetTemplate(ctx context.Context, req *ai_orchestrationv1.GetTemplateRequest) (*ai_orchestrationv1.GetTemplateResponse, error) {
	start := time.Now()

	store := globalTemplateStore
	store.mu.RLock()
	defer store.mu.RUnlock()

	tpl, ok := store.templates[req.GetTemplateId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "template %q not found", req.GetTemplateId())
	}

	classCount := 0
	propCount := 0
	indivCount := 0
	for _, step := range tpl.Steps {
		switch strings.ToUpper(step.Operation) {
		case "CREATE_CLASS":
			classCount++
		case "CREATE_OBJECT_PROPERTY", "CREATE_DATATYPE_PROPERTY":
			propCount++
		case "CREATE_INDIVIDUAL":
			indivCount++
		}
	}

	protoSteps := make([]*ai_orchestrationv1.SequenceStep, len(tpl.Steps))
	for i, step := range tpl.Steps {
		protoSteps[i] = toProtoSequenceStep(step)
	}

	slog.Debug("handler.templates.get",
		"template_id", req.GetTemplateId(),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &ai_orchestrationv1.GetTemplateResponse{
		Template: &ai_orchestrationv1.OntologyTemplate{
			Id:              tpl.ID,
			Name:            tpl.Name,
			Description:     tpl.Description,
			Domain:          tpl.Domain,
			Steps:           protoSteps,
			ClassCount:      int32(classCount),
			PropertyCount:   int32(propCount),
			IndividualCount: int32(indivCount),
		},
	}, nil
}

// ListTemplates returns available ontology templates, optionally filtered by domain.
