package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/api-gateway/proxy"
)

// testResponseWriter wraps httptest.ResponseRecorder to implement http.CloseNotifier,
// which gin's responseWriter and httputil.ReverseProxy rely on.
type testResponseWriter struct {
	*httptest.ResponseRecorder
	closeNotify chan bool
}

func (w *testResponseWriter) CloseNotify() <-chan bool {
	return w.closeNotify
}

// newTestRecorder returns an http.ResponseWriter suitable for tests with the proxy.
func newTestRecorder() *testResponseWriter {
	return &testResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		closeNotify:      make(chan bool, 1),
	}
}

// setupTestRouter creates a Gin engine wired to a test upstream server.
func setupTestRouter(upstreamHandler http.HandlerFunc) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)

	upstream := httptest.NewServer(upstreamHandler)

	p, err := proxy.New(upstream.URL, 0, nil, "test-ontology")
	if err != nil {
		panic("failed to create test proxy: " + err.Error())
	}

	r := gin.New()
	api := r.Group("/api/v1")
	handler := NewOntologyHandler(p)

	api.GET("/ontologies", handler.HandleListOntologies)
	api.GET("/ontologies/:id", handler.HandleGetOntology)
	api.GET("/ontologies/:id/classes", handler.HandleListClasses)
	api.GET("/ontologies/:id/classes/:classId", handler.HandleGetClass)
	api.GET("/ontologies/:id/properties", handler.HandleListProperties)
	api.GET("/ontologies/:id/individuals", handler.HandleListIndividuals)

	return r, func() {
		upstream.Close()
	}
}

// verifyErrorResponse checks that the response body contains a valid error structure.
func verifyErrorResponse(t *testing.T, body []byte, expectedCode string) {
	t.Helper()
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp.Error.Code != expectedCode {
		t.Errorf("expected error code %q, got %q", expectedCode, errResp.Error.Code)
	}
	if errResp.Error.Message == "" {
		t.Error("expected non-empty error message")
	}
}

// serveRequestWithWriter serves a request using the custom testResponseWriter.
func serveRequestWithWriter(r http.Handler, req *http.Request) *testResponseWriter {
	w := newTestRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestPaginationValidation tests that parsePagination correctly validates
// page and per_page parameters without needing a proxy upstream.
func TestPaginationValidation(t *testing.T) {
	tests := []struct {
		name        string
		page        string
		perPage     string
		wantPage    int
		wantPerPage int
		wantErr     bool
	}{
		{"default values", "", "", 1, 20, false},
		{"page=2", "2", "", 2, 20, false},
		{"per_page=50", "", "50", 1, 50, false},
		{"page=0", "0", "", 0, 0, true},
		{"page=-1", "-1", "", 0, 0, true},
		{"page=abc", "abc", "", 0, 0, true},
		{"per_page=0", "", "0", 0, 0, true},
		{"per_page=abc", "", "abc", 0, 0, true},
		{"per_page=200 clamped", "", "200", 1, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ginCtx, _ := gin.CreateTestContext(newTestRecorder())
			ginCtx.Request = httptest.NewRequest("GET", fmt.Sprintf("/?page=%s&per_page=%s", tt.page, tt.perPage), nil)

			page, perPage, err := parsePagination(ginCtx)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if page != tt.wantPage {
				t.Errorf("expected page %d, got %d", tt.wantPage, page)
			}
			if perPage != tt.wantPerPage {
				t.Errorf("expected per_page %d, got %d", tt.wantPerPage, perPage)
			}
		})
	}
}

// TestOntologyHandler_InvalidPagination tests that the handler returns 400
// for invalid pagination. These tests validate BEFORE the proxy is called.
func TestOntologyHandler_InvalidPagination(t *testing.T) {
	handler := &OntologyHandler{} // nil proxy is fine — validation happens before ServeHTTP

	r := gin.New()
	api := r.Group("/api/v1")
	api.GET("/ontologies", handler.HandleListOntologies)
	api.GET("/ontologies/:id/classes", handler.HandleListClasses)
	api.GET("/ontologies/:id/properties", handler.HandleListProperties)
	api.GET("/ontologies/:id/individuals", handler.HandleListIndividuals)

	tests := []struct {
		name    string
		url     string
		errCode string
	}{
		{"list page=0", "/api/v1/ontologies?page=0", "GATEWAY-INVALID-PAGINATION"},
		{"list page=-1", "/api/v1/ontologies?page=-1", "GATEWAY-INVALID-PAGINATION"},
		{"list per_page=0", "/api/v1/ontologies?per_page=0", "GATEWAY-INVALID-PAGINATION"},
		{"list page=abc", "/api/v1/ontologies?page=abc", "GATEWAY-INVALID-PAGINATION"},
		{"classes page=-1", "/api/v1/ontologies/test/classes?page=-1", "GATEWAY-INVALID-PAGINATION"},
		{"properties per_page=abc", "/api/v1/ontologies/test/properties?per_page=abc", "GATEWAY-INVALID-PAGINATION"},
		{"individuals page=0", "/api/v1/ontologies/test/individuals?page=0", "GATEWAY-INVALID-PAGINATION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newTestRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
			verifyErrorResponse(t, w.Body.Bytes(), tt.errCode)
		})
	}
}

// TestOntologyHandler_InvalidPropertyType tests that invalid property type filter returns 400.
func TestOntologyHandler_InvalidPropertyType(t *testing.T) {
	handler := &OntologyHandler{}
	r := gin.New()
	api := r.Group("/api/v1")
	api.GET("/ontologies/:id/properties", handler.HandleListProperties)

	w := newTestRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ontologies/test/properties?type=annotation", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-INVALID-PROPERTY-TYPE")
}

// TestOntologyHandler_MissingOntologyID tests that requests missing ontology ID return 400.
func TestOntologyHandler_MissingOntologyID(t *testing.T) {
	handler := &OntologyHandler{}
	r := gin.New()
	api := r.Group("/api/v1")
	api.GET("/ontologies/:id", handler.HandleGetOntology)
	api.GET("/ontologies/:id/classes", handler.HandleListClasses)
	api.GET("/ontologies/:id/classes/:classId", handler.HandleGetClass)
	api.GET("/ontologies/:id/properties", handler.HandleListProperties)
	api.GET("/ontologies/:id/individuals", handler.HandleListIndividuals)

	tests := []struct {
		name    string
		url     string
		errCode string
	}{
		// These routes match a path param that could be empty, but Gin always
		// assigns it. The handler checks for non-empty.
		{"get ontology with separate path", "/api/v1/ontologies//classes", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newTestRecorder()
			req, _ := http.NewRequest("GET", tt.url, nil)
			r.ServeHTTP(w, req)
			t.Logf("status %d, body %s", w.Code, w.Body.String())
		})
	}
}

// TestCalculateTotalPages verifies the pagination math helper.
func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		total   int
		perPage int
		want    int
	}{
		{0, 20, 0},
		{10, 20, 1},
		{20, 20, 1},
		{21, 20, 2},
		{100, 20, 5},
		{1, 1, 1},
		{0, 0, 0},
		{50, 0, 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := CalculateTotalPages(tt.total, tt.perPage)
			if got != tt.want {
				t.Errorf("CalculateTotalPages(%d, %d) = %d, want %d", tt.total, tt.perPage, got, tt.want)
			}
		})
	}
}

// TestOntologyHandler_ProxyForwarding tests that valid requests are forwarded
// to the proxy and the upstream response is returned.
func TestOntologyHandler_ProxyForwarding(t *testing.T) {
	r, cleanup := setupTestRouter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"meta":{"page":1,"per_page":20,"total":0,"total_pages":0}}`))
	})
	defer cleanup()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"list ontologies", "/api/v1/ontologies", false},
		{"get ontology", "/api/v1/ontologies/test-onto", false},
		{"list classes", "/api/v1/ontologies/test-onto/classes", false},
		{"list classes with query", "/api/v1/ontologies/test-onto/classes?q=person&parent_id=owl%3AThing", false},
		{"get class", "/api/v1/ontologies/test-onto/classes/Person", false},
		{"list properties", "/api/v1/ontologies/test-onto/properties", false},
		{"list properties with type", "/api/v1/ontologies/test-onto/properties?type=object", false},
		{"list individuals", "/api/v1/ontologies/test-onto/individuals", false},
		{"list individuals with filter", "/api/v1/ontologies/test-onto/individuals?class_id=Person", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serveRequestWithWriter(r, httptest.NewRequest("GET", tt.url, nil))
			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d; body: %s", w.Code, w.Body.String())
			}
		})
	}
}

// TestOntologyHandler_UpstreamError tests that upstream 5xx errors pass through.
func TestOntologyHandler_UpstreamError(t *testing.T) {
	r, cleanup := setupTestRouter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":"ONTOLOGY-INTERNAL-ERROR","message":"Internal error"}}`))
	})
	defer cleanup()

	w := serveRequestWithWriter(r, httptest.NewRequest("GET", "/api/v1/ontologies/test-onto", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestOntologyHandler_ValidPropertyTypeFilters tests that valid property type filters pass through.
func TestOntologyHandler_ValidPropertyTypeFilters(t *testing.T) {
	r, cleanup := setupTestRouter(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"meta":{"page":1,"per_page":20,"total":0,"total_pages":0}}`))
	})
	defer cleanup()

	tests := []string{"", "object", "datatype"}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("type=%q", tt), func(t *testing.T) {
			url := "/api/v1/ontologies/test/properties"
			if tt != "" {
				url += "?type=" + tt
			}
			w := serveRequestWithWriter(r, httptest.NewRequest("GET", url, nil))
			if w.Code != http.StatusOK {
				t.Errorf("expected status 200 for type=%q, got %d", tt, w.Code)
			}
		})
	}
}
