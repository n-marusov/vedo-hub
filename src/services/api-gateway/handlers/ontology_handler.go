package handlers

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/api-gateway/proxy"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

// OntologyHandler handles REST read endpoints for ontologies, classes,
// properties, and individuals. It validates query parameters, logs the
// request, and forwards to the ontology service via the reverse proxy.
type OntologyHandler struct {
	ontologyProxy *proxy.Proxy
}

// NewOntologyHandler creates a new OntologyHandler.
func NewOntologyHandler(ontologyProxy *proxy.Proxy) *OntologyHandler {
	return &OntologyHandler{
		ontologyProxy: ontologyProxy,
	}
}

// HandleListOntologies handles GET /api/v1/ontologies.
func (h *OntologyHandler) HandleListOntologies(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")

	page, perPage, err := parsePagination(c)
	if err != nil {
		slog.Warn("ontology.list.invalid_pagination",
			"trace_id", traceID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-INVALID-PAGINATION",
				Message: "Invalid pagination parameters. Page must be >= 1, per_page must be 1-100.",
			},
		})
		return
	}

	slog.Debug("ontology.list",
		"trace_id", traceID,
		"page", page,
		"per_page", perPage,
		"q", c.Query("q"),
		"user_id", c.GetHeader("X-User-Id"),
	)

	h.ontologyProxy.ServeHTTP(c.Writer, c.Request)

	slog.Info("ontology.list.completed",
		"trace_id", traceID,
		"page", page,
		"per_page", perPage,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleGetOntology handles GET /api/v1/ontologies/:id.
func (h *OntologyHandler) HandleGetOntology(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		slog.Warn("ontology.get.missing_id",
			"trace_id", traceID,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-MISSING-ONTOLOGY-ID",
				Message: "Ontology ID is required.",
			},
		})
		return
	}

	slog.Debug("ontology.get",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"user_id", c.GetHeader("X-User-Id"),
	)

	h.ontologyProxy.ServeHTTP(c.Writer, c.Request)

	slog.Info("ontology.get.completed",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleListClasses handles GET /api/v1/ontologies/:id/classes.
func (h *OntologyHandler) HandleListClasses(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	page, perPage, err := parsePagination(c)
	if err != nil {
		slog.Warn("ontology.classes.invalid_pagination",
			"trace_id", traceID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-INVALID-PAGINATION",
				Message: "Invalid pagination parameters. Page must be >= 1, per_page must be 1-100.",
			},
		})
		return
	}

	if ontologyID == "" {
		slog.Warn("ontology.classes.missing_id",
			"trace_id", traceID,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-MISSING-ONTOLOGY-ID",
				Message: "Ontology ID is required.",
			},
		})
		return
	}

	slog.Debug("ontology.classes.list",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"page", page,
		"per_page", perPage,
		"q", c.Query("q"),
		"parent_id", c.Query("parent_id"),
		"user_id", c.GetHeader("X-User-Id"),
	)

	h.ontologyProxy.ServeHTTP(c.Writer, c.Request)

	slog.Info("ontology.classes.list.completed",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"page", page,
		"per_page", perPage,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleGetClass handles GET /api/v1/ontologies/:id/classes/:classId.
func (h *OntologyHandler) HandleGetClass(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")
	classID := c.Param("classId")

	if ontologyID == "" || classID == "" {
		slog.Warn("ontology.class.get.missing_params",
			"trace_id", traceID,
			"ontology_id", ontologyID,
			"class_id", classID,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-MISSING-PARAM",
				Message: "Ontology ID and Class ID are required.",
			},
		})
		return
	}

	slog.Debug("ontology.class.get",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"class_id", classID,
		"user_id", c.GetHeader("X-User-Id"),
	)

	h.ontologyProxy.ServeHTTP(c.Writer, c.Request)

	slog.Info("ontology.class.get.completed",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"class_id", classID,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleListProperties handles GET /api/v1/ontologies/:id/properties.
func (h *OntologyHandler) HandleListProperties(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	page, perPage, err := parsePagination(c)
	if err != nil {
		slog.Warn("ontology.properties.invalid_pagination",
			"trace_id", traceID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-INVALID-PAGINATION",
				Message: "Invalid pagination parameters. Page must be >= 1, per_page must be 1-100.",
			},
		})
		return
	}

	if ontologyID == "" {
		slog.Warn("ontology.properties.missing_id",
			"trace_id", traceID,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-MISSING-ONTOLOGY-ID",
				Message: "Ontology ID is required.",
			},
		})
		return
	}

	propType := c.Query("type")
	if propType != "" && propType != "object" && propType != "datatype" {
		slog.Warn("ontology.properties.invalid_type",
			"trace_id", traceID,
			"type", propType,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-INVALID-PROPERTY-TYPE",
				Message: "Invalid property type filter. Must be 'object' or 'datatype'.",
			},
		})
		return
	}

	slog.Debug("ontology.properties.list",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"page", page,
		"per_page", perPage,
		"type", propType,
		"user_id", c.GetHeader("X-User-Id"),
	)

	h.ontologyProxy.ServeHTTP(c.Writer, c.Request)

	slog.Info("ontology.properties.list.completed",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"page", page,
		"per_page", perPage,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleListIndividuals handles GET /api/v1/ontologies/:id/individuals.
func (h *OntologyHandler) HandleListIndividuals(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	page, perPage, err := parsePagination(c)
	if err != nil {
		slog.Warn("ontology.individuals.invalid_pagination",
			"trace_id", traceID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-INVALID-PAGINATION",
				Message: "Invalid pagination parameters. Page must be >= 1, per_page must be 1-100.",
			},
		})
		return
	}

	if ontologyID == "" {
		slog.Warn("ontology.individuals.missing_id",
			"trace_id", traceID,
		)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-MISSING-ONTOLOGY-ID",
				Message: "Ontology ID is required.",
			},
		})
		return
	}

	slog.Debug("ontology.individuals.list",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"page", page,
		"per_page", perPage,
		"class_id", c.Query("class_id"),
		"user_id", c.GetHeader("X-User-Id"),
	)

	h.ontologyProxy.ServeHTTP(c.Writer, c.Request)

	slog.Info("ontology.individuals.list.completed",
		"trace_id", traceID,
		"ontology_id", ontologyID,
		"page", page,
		"per_page", perPage,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// parsePagination extracts and validates page and per_page query parameters.
func parsePagination(c *gin.Context) (page, perPage int, err error) {
	page = defaultPage
	perPage = defaultPerPage

	if p := c.Query("page"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 {
			return 0, 0, strconv.ErrSyntax
		}
		page = v
	}

	if pp := c.Query("per_page"); pp != "" {
		v, err := strconv.Atoi(pp)
		if err != nil || v < 1 {
			return 0, 0, strconv.ErrSyntax
		}
		if v > maxPerPage {
			v = maxPerPage
		}
		perPage = v
	}

	return page, perPage, nil
}

// CalculateTotalPages returns the total number of pages given total results and page size.
func CalculateTotalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(perPage)))
}
