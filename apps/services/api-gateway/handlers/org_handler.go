package handlers

// OrgHandler provides HTTP handlers for organizational model REST endpoints.
// Maps Gin HTTP requests to gRPC org service calls and returns JSON responses.
//
// Route table:
//   GET    /api/v1/groups              → HandleListGroups
//   POST   /api/v1/groups              → HandleCreateGroup
//   GET    /api/v1/groups/:id          → HandleGetGroup
//   PUT    /api/v1/groups/:id          → HandleUpdateGroup
//   DELETE /api/v1/groups/:id          → HandleDeleteGroup
//   GET    /api/v1/groups/:id/subgroups → HandleListChildGroups
//   GET    /api/v1/projects            → HandleListProjects
//   POST   /api/v1/projects            → HandleCreateProject
//   GET    /api/v1/projects/:id        → HandleGetProject
//   PUT    /api/v1/projects/:id        → HandleUpdateProject
//   DELETE /api/v1/projects/:id        → HandleDeleteProject
//   POST   /api/v1/projects/:id/fork   → HandleForkProject
//   GET    /api/v1/projects/:id/members         → HandleListMembers
//   POST   /api/v1/projects/:id/members         → HandleAddMember
//   PUT    /api/v1/projects/:id/members/:userId → HandleUpdateMemberRole
//   DELETE /api/v1/projects/:id/members/:userId → HandleRemoveMember
//   GET    /api/v1/groups/:id/members             → HandleListMembers
//   PUT    /api/v1/projects/:id/visibility      → HandleSetVisibility
//   GET    /api/v1/projects/:id/visibility      → HandleGetVisibility
//   GET    /api/v1/projects/:id/policies        → HandleListPolicies
//   POST   /api/v1/projects/:id/policies        → HandleCreatePolicy
//   DELETE /api/v1/projects/:id/policies/:policyId → HandleDeletePolicy

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/api-gateway/proxy"
	authv1 "vedo-core/src/services/shared/proto/auth/v1"
)

// OrgHandler handles org management HTTP endpoints via gRPC backend calls.
type OrgHandler struct {
	orgClient *proxy.OrgServiceClient
}

// NewOrgHandler creates a new OrgHandler with the given gRPC org client.
func NewOrgHandler(orgClient *proxy.OrgServiceClient) *OrgHandler {
	return &OrgHandler{orgClient: orgClient}
}

// extractToken extracts the Bearer token from the Authorization header.
func extractToken(c *gin.Context) string {
	return c.GetHeader("Authorization")
}

// ============================================================================
// Group Handlers
// ============================================================================

func (h *OrgHandler) HandleListGroups(c *gin.Context) {
	token := extractToken(c)

	resp, err := h.orgClient.ListGroups(c.Request.Context(), &authv1.ListGroupsRequest{
		Search: c.Query("search"),
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetGroups()})
}

func (h *OrgHandler) HandleCreateGroup(c *gin.Context) {
	token := extractToken(c)

	var req struct {
		Name        string `json:"name"`
		Label       string `json:"label"` // deprecated — use name
		Description string `json:"description"`
		ParentID    string `json:"parent_id"`
		Visibility  string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("http.org.create_group.invalid_request", "error", err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	// Canonical field is "name"; accept legacy "label" as fallback.
	groupName := req.Name
	if groupName == "" && req.Label != "" {
		groupName = req.Label
		slog.Warn("http.org.deprecated_label_name",
			"label", req.Label,
		)
	}

	slog.Info("http.org.create_group.request",
		"name", groupName,
		"visibility", req.Visibility,
		"parent_id", req.ParentID,
	)

	resp, err := h.orgClient.CreateGroup(c.Request.Context(), &authv1.CreateGroupRequest{
		Name:        groupName,
		Description: req.Description,
		ParentId:    req.ParentID,
		Visibility:  req.Visibility,
	}, token)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			httpCode, errCode := mapGrpcCodeToHTTP(st.Code())
			slog.Error("http.org.create_group.grpc_error",
				"grpc_code", st.Code(),
				"http_code", httpCode,
				"error", st.Message(),
			)
			c.JSON(httpCode, models.ErrorResponse{
				Error: models.ErrorDetail{Code: errCode, Message: st.Message()},
			})
		} else {
			slog.Error("http.org.create_group.error", "error", err.Error())
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
			})
		}
		return
	}

	group := resp.GetGroup()
	slog.Info("http.org.create_group.success",
		"group_id", group.GetId(),
		"name", groupName,
	)
	c.JSON(http.StatusCreated, gin.H{"data": group})
}

// mapGrpcCodeToHTTP maps a gRPC status code to an HTTP status code and error code string.
func mapGrpcCodeToHTTP(grpcCode codes.Code) (int, string) {
	switch grpcCode {
	case codes.PermissionDenied:
		return http.StatusForbidden, "FORBIDDEN"
	case codes.NotFound:
		return http.StatusNotFound, "NOT_FOUND"
	case codes.InvalidArgument:
		return http.StatusBadRequest, "INVALID_ARGUMENT"
	case codes.AlreadyExists:
		return http.StatusConflict, "CONFLICT"
	case codes.FailedPrecondition:
		return http.StatusBadRequest, "FAILED_PRECONDITION"
	default:
		return http.StatusInternalServerError, "INTERNAL"
	}
}

func (h *OrgHandler) HandleGetGroup(c *gin.Context) {
	token := extractToken(c)
	id := c.Param("id")

	resp, err := h.orgClient.GetGroup(c.Request.Context(), id, token)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "NOT_FOUND", Message: "Group not found"},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetGroup()})
}

func (h *OrgHandler) HandleUpdateGroup(c *gin.Context) {
	token := extractToken(c)
	id := c.Param("id")

	var req struct {
		Label       string `json:"label"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	slog.Debug("http.org.update_group",
		"id", id,
		"name", req.Label,
		"visibility", req.Visibility,
	)

	resp, err := h.orgClient.UpdateGroup(c.Request.Context(), &authv1.UpdateGroupRequest{
		Id:          id,
		Name:        req.Label,
		Description: req.Description,
		Visibility:  req.Visibility,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetGroup()})
}

func (h *OrgHandler) HandleDeleteGroup(c *gin.Context) {
	token := extractToken(c)
	id := c.Param("id")

	_, err := h.orgClient.DeleteGroup(c.Request.Context(), id, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrgHandler) HandleListChildGroups(c *gin.Context) {
	token := extractToken(c)
	parentID := c.Param("id")

	resp, err := h.orgClient.ListChildGroups(c.Request.Context(), parentID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetGroups()})
}

// ============================================================================
// Project Handlers
// ============================================================================

func (h *OrgHandler) HandleListProjects(c *gin.Context) {
	token := extractToken(c)

	resp, err := h.orgClient.ListProjects(c.Request.Context(), &authv1.ListProjectsRequest{
		Search:  c.Query("search"),
		SortBy:  c.Query("sortBy"),
		SortDir: c.Query("sortDir"),
		Page:    int32(atoi(c.Query("page"), 1)),
		PerPage: int32(atoi(c.Query("perPage"), 10)),
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetProjects(), "total": resp.GetTotal()})
}

func (h *OrgHandler) HandleCreateProject(c *gin.Context) {
	token := extractToken(c)

	var req struct {
		Name        string `json:"name"`
		Label       string `json:"label"` // deprecated — use name
		Description string `json:"description"`
		GroupID     string `json:"group_id"`
		Visibility  string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("http.org.create_project.invalid_request", "error", err.Error())
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	// Canonical field is "name"; accept legacy "label" as fallback.
	projectName := req.Name
	if projectName == "" && req.Label != "" {
		projectName = req.Label
		slog.Warn("http.org.deprecated_label_project_name",
			"label", req.Label,
		)
	}

	slog.Info("http.org.create_project.request",
		"name", projectName,
		"group_id", req.GroupID,
		"visibility", req.Visibility,
	)

	resp, err := h.orgClient.CreateProject(c.Request.Context(), &authv1.CreateProjectRequest{
		Name:        projectName,
		Description: req.Description,
		GroupId:     req.GroupID,
		Visibility:  req.Visibility,
	}, token)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			httpCode, errCode := mapGrpcCodeToHTTP(st.Code())
			slog.Error("http.org.create_project.grpc_error",
				"grpc_code", st.Code(),
				"http_code", httpCode,
				"error", st.Message(),
			)
			c.JSON(httpCode, models.ErrorResponse{
				Error: models.ErrorDetail{Code: errCode, Message: st.Message()},
			})
		} else {
			slog.Error("http.org.create_project.error", "error", err.Error())
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
			})
		}
		return
	}

	project := resp.GetProject()
	slog.Info("http.org.create_project.success",
		"project_id", project.GetId(),
		"name", projectName,
		"visibility", project.GetVisibility(),
		"ontology_id", project.GetOntologyId(),
	)
	c.JSON(http.StatusCreated, gin.H{"data": project})
}

func (h *OrgHandler) HandleGetProject(c *gin.Context) {
	token := extractToken(c)
	id := c.Param("id")

	resp, err := h.orgClient.GetProject(c.Request.Context(), id, token)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "NOT_FOUND", Message: "Project not found"},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetProject()})
}

func (h *OrgHandler) HandleUpdateProject(c *gin.Context) {
	token := extractToken(c)
	id := c.Param("id")

	var req struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	resp, err := h.orgClient.UpdateProject(c.Request.Context(), &authv1.UpdateProjectRequest{
		Id:          id,
		Name:        req.Label,
		Description: req.Description,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetProject()})
}

func (h *OrgHandler) HandleDeleteProject(c *gin.Context) {
	token := extractToken(c)
	id := c.Param("id")

	_, err := h.orgClient.DeleteProject(c.Request.Context(), id, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrgHandler) HandleForkProject(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	sourceProjectID := c.Param("id")
	token := extractToken(c)

	resp, err := h.orgClient.ForkProject(c.Request.Context(), &authv1.ForkProjectRequest{
		SourceProjectId: sourceProjectID,
	}, token)
	if err != nil {
		slog.Error("fork.endpoint.failed", "source", sourceProjectID, "trace_id", traceID, "error", err)
		// Map gRPC error to HTTP status
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.PermissionDenied, codes.Unauthenticated:
				c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error: models.ErrorDetail{Code: "FORBIDDEN", Message: "No read access to source project"},
				})
				return
			case codes.NotFound:
				c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error: models.ErrorDetail{Code: "NOT_FOUND", Message: "Source project not found"},
				})
				return
			case codes.Unavailable:
				c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
					Error: models.ErrorDetail{Code: "SERVICE_UNAVAILABLE", Message: "Service temporarily unavailable, please retry"},
				})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}

	proj := resp.GetProject()
	c.JSON(http.StatusCreated, gin.H{
		"project_id":          proj.GetId(),
		"ontology_id":         proj.GetOntologyId(),
		"upstream_project_id": proj.GetUpstreamProjectId(),
	})

	slog.Info("fork.endpoint.completed", "source", sourceProjectID, "project_id", proj.GetId(), "trace_id", traceID, "duration_ms", time.Since(start).Milliseconds())
}

// HandleMoveProject transfers a Project from one Group to another.
//
// PUT /api/v1/projects/:id/move  { target_group_id: "..." }
//
// Idempotency-Key is required (enforced by middleware).
// RBAC: Only Owner of both source and target groups can move.
// Audit: emits project.moved event with object_id, target_group_id.
//
// TODO (Task 3.1): Uncomment gRPC path once `buf generate` regenerates
// protobuf Go types for MoveProjectRequest/Response. For now, returns 501
// Not Implemented.
func (h *OrgHandler) HandleMoveProject(c *gin.Context) {
	traceID := c.GetHeader("X-Trace-Id")
	projectID := c.Param("id")

	slog.Warn("move.endpoint.not_implemented",
		"project_id", projectID,
		"trace_id", traceID,
		"hint", "awaiting buf generate for MoveProject proto types",
	)

	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    "NOT_IMPLEMENTED",
			Message: "Project move endpoint is pending protobuf code generation. Run `buf generate` and wire the gRPC path.",
		},
	})

	// TODO: gRPC path (uncomment after buf generate):
	// token := extractToken(c)
	// var req struct {
	// 	TargetGroupID string `json:"target_group_id"`
	// }
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, models.ErrorResponse{
	// 		Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
	// 	})
	// 	return
	// }
	// resp, err := h.orgClient.MoveProject(c.Request.Context(), &authv1.MoveProjectRequest{
	// 	ProjectId:     projectID,
	// 	TargetGroupId: req.TargetGroupID,
	// }, token)
	// ... (map gRPC error codes to HTTP, same pattern as ForkProject)
	// c.JSON(http.StatusOK, gin.H{"project": resp.GetProject()})
}

// ============================================================================
// Member Handlers
// ============================================================================

func (h *OrgHandler) HandleListMembers(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")

	resp, err := h.orgClient.ListMembers(c.Request.Context(), scope, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetMembers()})
}

func (h *OrgHandler) HandleAddMember(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")

	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	resp, err := h.orgClient.AddMember(c.Request.Context(), &authv1.AddMemberRequest{
		Scope:  scope,
		UserId: req.UserID,
		Role:   req.Role,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp.GetMember()})
}

func (h *OrgHandler) HandleUpdateMemberRole(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")
	userID := c.Param("userId")

	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	resp, err := h.orgClient.UpdateMemberRole(c.Request.Context(), &authv1.UpdateMemberRoleRequest{
		Scope:  scope,
		UserId: userID,
		Role:   req.Role,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetMember()})
}

func (h *OrgHandler) HandleRemoveMember(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")
	userID := c.Param("userId")

	_, err := h.orgClient.RemoveMember(c.Request.Context(), &authv1.RemoveMemberRequest{
		Scope:  scope,
		UserId: userID,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// ============================================================================
// Visibility Handlers
// ============================================================================

func (h *OrgHandler) HandleSetVisibility(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")

	var req struct {
		Visibility string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	_, err := h.orgClient.SetVisibility(c.Request.Context(), &authv1.SetVisibilityRequest{
		Scope:      scope,
		Visibility: req.Visibility,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *OrgHandler) HandleGetVisibility(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")

	resp, err := h.orgClient.GetVisibility(c.Request.Context(), scope, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"visibility": resp.GetVisibility()})
}

// ============================================================================
// Policy Handlers
// ============================================================================

func (h *OrgHandler) HandleListPolicies(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")

	resp, err := h.orgClient.ListPolicies(c.Request.Context(), scope, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp.GetPolicies()})
}

func (h *OrgHandler) HandleCreatePolicy(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")

	var req struct {
		Pattern map[string]string `json:"pattern"`
		Right   string            `json:"right"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	_, err := h.orgClient.CreatePolicy(c.Request.Context(), &authv1.CreatePolicyRequest{
		Scope:   scope,
		Pattern: req.Pattern,
		Right:   req.Right,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (h *OrgHandler) HandleDeletePolicy(c *gin.Context) {
	token := extractToken(c)
	scope := c.Param("id")
	policyID := c.Param("policyId")

	_, err := h.orgClient.DeletePolicy(c.Request.Context(), &authv1.DeletePolicyRequest{
		Scope:    scope,
		PolicyId: policyID,
	}, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "GRPC_ERROR", Message: err.Error()},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// atoi parses an int from a string, returning the default on error.
func atoi(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return defaultVal
		}
		n = n*10 + int(c-'0')
	}
	return n
}
