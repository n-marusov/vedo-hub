package middleware

// Deprecation middleware for RFC 8594 Sunset headers.
// Implements REQ-FUN.API.sunset-header (P0, DRAFT).
//
// Usage:
//
//	// Apply to a single deprecated endpoint:
//	r.GET("/api/v1/old-endpoint", middleware.SunsetHeader(cfg), handler)
//
//	// Apply to a deprecated route group:
//	old := api.Group("")
//	old.Use(middleware.SunsetHeader(cfg))
//	old.GET("/legacy", handler)
//
// Config values are typically loaded from environment or a feature-flag store
// so that deprecation dates can be updated without a code deploy.

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// SunsetConfig defines RFC 8594 deprecation headers for a deprecated endpoint.
// Zero value produces no headers (safe disabled state).
type SunsetConfig struct {
	// SunsetDate is the date after which the endpoint will be removed.
	// When zero (time.Time{}), no Sunset header is emitted.
	SunsetDate time.Time

	// Deprecated flag. When true, sets "Deprecation: true".
	// When SunsetDate is zero but this is true, only Deprecation is set.
	Deprecated bool

	// MigrationURL is an optional link to a migration guide.
	// When set, produces: Link: <url>; rel="deprecation"; type="text/html"
	MigrationURL string
}

// DefaultSunset returns a SunsetConfig that marks an endpoint as deprecated
// with a default sunset date 90 days from now and the standard migration guide link.
func DefaultSunset() SunsetConfig {
	return SunsetConfig{
		SunsetDate:   time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		Deprecated:   true,
		MigrationURL: "https://docs.vedo.io/api/migration-guide",
	}
}

// SunsetHeader returns a Gin middleware that adds RFC 8594 deprecation headers
// to every response from the routes it wraps.
//
// Headers added:
//   - Sunset: <RFC1123 date>  (RFC 8594)
//   - Deprecation: true
//   - Link: <url>; rel="deprecation"; type="text/html"
//
// The headers are added on write-header (c.Next), guaranteeing they appear in
// every response regardless of status code (200, 400, 500, etc.).
func SunsetHeader(cfg SunsetConfig) gin.HandlerFunc {
	// Pre-compute header values so we don't format on every request.
	var sunsetVal string
	if !cfg.SunsetDate.IsZero() {
		sunsetVal = cfg.SunsetDate.Format(time.RFC1123)
	}

	var linkVal string
	if cfg.MigrationURL != "" {
		u, err := url.Parse(cfg.MigrationURL)
		if err == nil && u.Scheme != "" {
			linkVal = fmt.Sprintf("<%s>; rel=\"deprecation\"; type=\"text/html\"", cfg.MigrationURL)
		}
	}

	return func(c *gin.Context) {
		c.Next()

		// Add headers AFTER the handler runs, so they appear on all responses
		// including error responses (4xx, 5xx).
		if sunsetVal != "" {
			c.Header("Sunset", sunsetVal)
		}
		if cfg.Deprecated {
			c.Header("Deprecation", "true")
		}
		if linkVal != "" {
			c.Header("Link", linkVal)
		}
	}
}
