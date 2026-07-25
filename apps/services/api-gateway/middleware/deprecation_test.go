package middleware

// Tests for RFC 8594 Sunset header middleware.
// BDD: [Condition]_[Action]_[ExpectedResult]
//
// Validates: REQ-FUN.API.sunset-header

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// setupSunsetTest creates a Gin engine with the SunsetHeader middleware
// applied to a /test route that returns 200 with a JSON body.
func setupSunsetTest(cfg SunsetConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", SunsetHeader(cfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

// setupSunsetTestWithError creates a Gin engine where the handler returns
// an error status, to verify headers appear on non-200 responses.
func setupSunsetTestWithError(cfg SunsetConfig, status int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", SunsetHeader(cfg), func(c *gin.Context) {
		c.JSON(status, gin.H{"error": "something went wrong"})
	})
	return r
}

func TestSunsetHeader_DefaultConfig_SetsAllHeaders(t *testing.T) {
	// BDD: [Default config]_[returns 200]_[sets Sunset, Deprecation, and Link headers]
	cfg := DefaultSunset()
	r := setupSunsetTest(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if w.Header().Get("Sunset") == "" {
		t.Error("expected Sunset header to be set")
	}
	if w.Header().Get("Deprecation") != "true" {
		t.Errorf("expected Deprecation: true, got %q", w.Header().Get("Deprecation"))
	}
	if w.Header().Get("Link") == "" {
		t.Error("expected Link header to be set")
	}

	// Verify Sunset header is a valid RFC1123 date
	_, err := time.Parse(time.RFC1123, w.Header().Get("Sunset"))
	if err != nil {
		t.Errorf("Sunset header is not a valid RFC1123 date: %v", err)
	}
}

func TestSunsetHeader_DefaultConfig_SetsHeadersOnErrorResponse(t *testing.T) {
	// BDD: [Default config]_[returns 400]_[still sets all headers]
	cfg := DefaultSunset()
	r := setupSunsetTestWithError(cfg, http.StatusBadRequest)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	if w.Header().Get("Sunset") == "" {
		t.Error("expected Sunset header on error response")
	}
	if w.Header().Get("Deprecation") != "true" {
		t.Error("expected Deprecation header on error response")
	}
}

func TestSunsetHeader_DefaultConfig_SetsHeadersOnServerError(t *testing.T) {
	// BDD: [Default config]_[returns 500]_[still sets all headers]
	cfg := DefaultSunset()
	r := setupSunsetTestWithError(cfg, http.StatusInternalServerError)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	if w.Header().Get("Sunset") == "" {
		t.Error("expected Sunset header on 500 response")
	}
	if w.Header().Get("Deprecation") != "true" {
		t.Error("expected Deprecation header on 500 response")
	}
}

func TestSunsetHeader_ZeroConfig_NoHeaders(t *testing.T) {
	// BDD: [Zero config]_[returns 200]_[no headers set]
	cfg := SunsetConfig{} // zero value — disabled
	r := setupSunsetTest(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("Sunset") != "" {
		t.Error("expected no Sunset header for zero config")
	}
	if w.Header().Get("Deprecation") != "" {
		t.Error("expected no Deprecation header for zero config")
	}
	if w.Header().Get("Link") != "" {
		t.Error("expected no Link header for zero config")
	}
}

func TestSunsetHeader_DeprecatedOnly_NoSunsetDate(t *testing.T) {
	// BDD: [Deprecated=true but no date]_[returns 200]_[sets Deprecation but not Sunset]
	cfg := SunsetConfig{
		Deprecated: true,
	}
	r := setupSunsetTest(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("Sunset") != "" {
		t.Error("expected no Sunset header when SunsetDate is zero")
	}
	if w.Header().Get("Deprecation") != "true" {
		t.Errorf("expected Deprecation: true, got %q", w.Header().Get("Deprecation"))
	}
}

func TestSunsetHeader_CustomDate_ReturnsFormattedHeader(t *testing.T) {
	// BDD: [Custom sunset date]_[returns 200]_[Sunset header matches RFC1123]
	customDate := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := SunsetConfig{
		SunsetDate: customDate,
		Deprecated: true,
	}
	r := setupSunsetTest(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	expected := "Fri, 01 Jan 2027 00:00:00 UTC"
	if w.Header().Get("Sunset") != expected {
		t.Errorf("expected Sunset %q, got %q", expected, w.Header().Get("Sunset"))
	}
}

func TestSunsetHeader_CustomMigrationURL_ReturnsInLinkHeader(t *testing.T) {
	// BDD: [Custom migration URL]_[returns 200]_[Link header contains URL]
	cfg := SunsetConfig{
		Deprecated:   true,
		MigrationURL: "https://example.com/migration-guide",
	}
	r := setupSunsetTest(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	expected := "<https://example.com/migration-guide>; rel=\"deprecation\"; type=\"text/html\""
	got := w.Header().Get("Link")
	if got != expected {
		t.Errorf("expected Link %q, got %q", expected, got)
	}
}

func TestSunsetHeader_InvalidMigrationURL_NoLinkHeader(t *testing.T) {
	// BDD: [Invalid migration URL]_[returns 200]_[no Link header emitted]
	cfg := SunsetConfig{
		Deprecated:   true,
		MigrationURL: "not-a-valid-url", // missing scheme
	}
	r := setupSunsetTest(cfg)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("Link") != "" {
		t.Error("expected no Link header for invalid migration URL")
	}
}

func TestSunsetHeader_AppliedToRouteGroup_HeadersOnAllRoutes(t *testing.T) {
	// BDD: [Middleware on route group]_[multiple routes]_[headers on all of them]
	gin.SetMode(gin.TestMode)
	cfg := DefaultSunset()
	r := gin.New()
	deprecated := r.Group("/old")
	deprecated.Use(SunsetHeader(cfg))
	deprecated.GET("/a", func(c *gin.Context) { c.Status(http.StatusOK) })
	deprecated.GET("/b", func(c *gin.Context) { c.Status(http.StatusOK) })

	for _, path := range []string{"/old/a", "/old/b"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)

		if w.Header().Get("Sunset") == "" {
			t.Errorf("expected Sunset header on %s", path)
		}
		if w.Header().Get("Deprecation") != "true" {
			t.Errorf("expected Deprecation header on %s", path)
		}
	}
}

func TestSunsetHeader_NotApplied_RouteHasNoHeaders(t *testing.T) {
	// BDD: [Route without middleware]_[returns 200]_[no sunset headers]
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/active", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/active", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("Sunset") != "" {
		t.Error("expected no Sunset header on non-deprecated route")
	}
	if w.Header().Get("Deprecation") != "" {
		t.Error("expected no Deprecation header on non-deprecated route")
	}
}

func TestSunsetHeader_ConcurrentRequests_NoRace(t *testing.T) {
	// BDD: [Concurrent requests]_[middleware active]_[no data race]
	cfg := DefaultSunset()
	r := setupSunsetTest(cfg)

	for i := 0; i < 20; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			r.ServeHTTP(w, req)
		}()
	}
}
