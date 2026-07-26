package proxy

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcUnaryHandler is the signature for a gRPC unary call that the
// GrpcHandler adapter wraps into a Gin HTTP handler.
type GrpcUnaryHandler[Req any, Resp any] func(context.Context, *Req) (*Resp, error)

// GrpcHandler adapts a gRPC unary call into a Gin HTTP handler.
// It reads the request body as JSON (mapped to Req), calls the gRPC method,
// and writes the proto response as JSON to the HTTP response.
// Path parameters (e.g., :id) are injected into the request if the Req type
// has SetOntologyId or similar setter methods — see requestInjectors.
func GrpcHandler[Req any, Resp any](
	grpcFunc func(context.Context, *Req) (*Resp, error),
	pathParams ...PathParamInjector,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		traceID := c.GetHeader("X-Trace-Id")

		var req Req
		if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
			slog.Warn("grpc_handler.bind_error",
				"trace_id", traceID,
				"error", err,
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "GATEWAY-INVALID-REQUEST", "message": "Invalid request body."},
			})
			return
		}

		// Inject path parameters into the request
		for _, injector := range pathParams {
			injector(c, &req)
		}

		ctx := c.Request.Context()
		resp, err := grpcFunc(ctx, &req)
		if err != nil {
			slog.Warn("grpc_handler.call_error",
				"trace_id", traceID,
				"error", err,
				"duration_ms", time.Since(start).Milliseconds(),
			)
			grpcErrToHTTP(c, err)
			return
		}

		slog.Debug("grpc_handler.ok",
			"trace_id", traceID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		c.JSON(http.StatusOK, resp)
	}
}

// PathParamInjector injects a Gin path parameter value into a gRPC request.
type PathParamInjector func(c *gin.Context, req interface{})

// WithOntologyID returns a PathParamInjector that sets the ontology_id field
// on request types that have a SetOntologyId method.
func WithOntologyID(paramName string) PathParamInjector {
	return func(c *gin.Context, req interface{}) {
		val := c.Param(paramName)
		if val == "" {
			return
		}
		if setter, ok := req.(interface{ SetOntologyId(string) }); ok {
			setter.SetOntologyId(val)
		}
	}
}

// grpcErrToHTTP converts a gRPC status error to an HTTP JSON error response.
func grpcErrToHTTP(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "GATEWAY-INTERNAL", "message": "Internal server error"},
		})
		return
	}

	var httpCode int
	var errCode string

	switch st.Code() {
	case codes.InvalidArgument:
		httpCode = http.StatusBadRequest
		errCode = "GATEWAY-INVALID-REQUEST"
	case codes.NotFound:
		httpCode = http.StatusNotFound
		errCode = "GATEWAY-NOT-FOUND"
	case codes.PermissionDenied:
		httpCode = http.StatusForbidden
		errCode = "LLM-POLICY-BLOCKED"
	case codes.Unavailable:
		httpCode = http.StatusServiceUnavailable
		errCode = "GATEWAY-LLM-UNAVAILABLE"
	case codes.Unimplemented:
		httpCode = http.StatusNotImplemented
		errCode = "GATEWAY-NOT-IMPLEMENTED"
	case codes.FailedPrecondition:
		httpCode = http.StatusBadRequest
		errCode = "GATEWAY-FAILED-PRECONDITION"
	default:
		httpCode = http.StatusInternalServerError
		errCode = "GATEWAY-INTERNAL"
	}

	c.JSON(httpCode, gin.H{
		"error": gin.H{
			"code":    errCode,
			"message": st.Message(),
		},
	})
}
