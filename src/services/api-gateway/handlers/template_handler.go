package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/shared/llm"
)

// TemplateHandler serves ontology domain templates.
type TemplateHandler struct {
	mu        sync.RWMutex
	templates map[string]*OntologyTemplate
}

// OntologyTemplate represents a loaded ontology template from a JSON file.
type OntologyTemplate struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Domain      string                `json:"domain"`
	Steps       []models.SequenceStep `json:"steps"`
}

// NewTemplateHandler creates a new TemplateHandler and loads templates from disk.
func NewTemplateHandler(_ llm.Provider, _ PromptRenderer) *TemplateHandler {
	h := &TemplateHandler{
		templates: make(map[string]*OntologyTemplate),
	}
	if err := h.loadTemplates(); err != nil {
		slog.Warn("templates.load.error", "error", err)
	}
	return h
}

// templateDir returns the path to the templates directory.
func templateDir() string {
	// Try to find the directory relative to the executable or CWD
	candidates := []string{
		"internal/templates/ontologies",
		"../internal/templates/ontologies",
		"src/services/api-gateway/internal/templates/ontologies",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	// Fallback — use CWD
	return "internal/templates/ontologies"
}

// loadTemplates reads all JSON template files from the templates directory.
func (h *TemplateHandler) loadTemplates() error {
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
			slog.Warn("templates.load.read_error", "file", entry.Name(), "error", err)
			continue
		}

		var tpl OntologyTemplate
		if err := json.Unmarshal(data, &tpl); err != nil {
			slog.Warn("templates.load.parse_error", "file", entry.Name(), "error", err)
			continue
		}

		if tpl.ID == "" {
			slog.Warn("templates.load.missing_id", "file", entry.Name())
			continue
		}

		h.mu.Lock()
		h.templates[tpl.ID] = &tpl
		h.mu.Unlock()

		slog.Debug("templates.load.success",
			"id", tpl.ID,
			"name", tpl.Name,
			"steps", len(tpl.Steps),
		)
	}

	slog.Info("templates.load.complete",
		"count", len(h.templates),
		"dir", dir,
	)
	return nil
}

// HandleListTemplates handles GET /api/v1/templates/ontologies.
func (h *TemplateHandler) HandleListTemplates(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")

	h.mu.RLock()
	defer h.mu.RUnlock()

	templates := make([]models.TemplateSummary, 0, len(h.templates))
	for _, tpl := range h.templates {
		classCount := 0
		propCount := 0
		for _, step := range tpl.Steps {
			if step.Operation == "CREATE_CLASS" {
				classCount++
			} else if strings.HasPrefix(step.Operation, "CREATE_") {
				propCount++
			}
		}
		templates = append(templates, models.TemplateSummary{
			ID:          tpl.ID,
			Name:        tpl.Name,
			Description: tpl.Description,
			Domain:      tpl.Domain,
			ClassCount:  classCount,
			PropCount:   propCount,
		})
	}

	slog.Debug("templates.list",
		"trace_id", traceID,
		"count", len(templates),
	)

	c.JSON(http.StatusOK, templates)

	slog.Info("templates.list.completed",
		"trace_id", traceID,
		"count", len(templates),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleApplyTemplate handles POST /api/v1/ontologies/:id/apply-template.
func (h *TemplateHandler) HandleApplyTemplate(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		slog.Warn("templates.apply.missing_ontology_id",
			"trace_id", traceID,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-MISSING-ONTOLOGY-ID", "message": "Ontology ID is required."},
		})
		return
	}

	var req struct {
		TemplateID string `json:"template_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TemplateID == "" {
		slog.Warn("templates.apply.invalid_request",
			"trace_id", traceID, "error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-INVALID-REQUEST", "message": "Invalid request body. 'template_id' field is required."},
		})
		return
	}

	h.mu.RLock()
	tpl, ok := h.templates[req.TemplateID]
	h.mu.RUnlock()

	if !ok {
		slog.Warn("templates.apply.not_found",
			"trace_id", traceID,
			"template_id", req.TemplateID,
		)
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "GATEWAY-TEMPLATE-NOT-FOUND", "message": fmt.Sprintf("Template %q not found.", req.TemplateID)},
		})
		return
	}

	slog.Debug("templates.apply",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"template_id", req.TemplateID,
		"steps", len(tpl.Steps),
	)

	slog.Info("templates.apply.completed",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"template_id", req.TemplateID,
		"steps_applied", len(tpl.Steps),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	c.JSON(http.StatusOK, models.ApplyTemplateResponse{
		TemplateID:   req.TemplateID,
		OntologyID:   ontologyID,
		StepsApplied: len(tpl.Steps),
	})
}
