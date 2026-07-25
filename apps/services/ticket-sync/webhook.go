// @ctx: Inbound webhook with HMAC-SHA256 validation shell
// @hlv:sec [CRYPTO] — HMAC-SHA256 signature verified on every webhook request

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

const webhookSecretKey = "webhook-hmac-secret"

func WebhookHMACMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetString("trace_id")
		slog.Info("webhook.hmac.enter",
			"trace_id", traceID,
		)

		// @hlv:sec [CRYPTO] — extract and validate signature header
		signature := c.GetHeader("X-Gitlab-Token")
		if signature == "" {
			slog.Error("webhook.hmac.missing_signature",
				"trace_id", traceID,
			)
			// @hlv SYNC-UNAUTHORIZED
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   ErrUnauthorized,
				Message: "Missing HMAC signature",
				Ref:     "security considerations",
			})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("webhook.hmac.read_error",
				"trace_id", traceID,
				"error", err.Error(),
			)
			// @hlv SYNC-UNAUTHORIZED
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   ErrUnauthorized,
				Message: "Failed to read request body",
				Ref:     "security considerations",
			})
			return
		}

		// @hlv:sec [CRYPTO] — compute expected HMAC and compare
		mac := hmac.New(sha256.New, []byte(webhookSecretKey))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			slog.Error("webhook.hmac.invalid_signature",
				"trace_id", traceID,
			)
			// @hlv SYNC-UNAUTHORIZED
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   ErrUnauthorized,
				Message: "Invalid HMAC signature",
				Ref:     "security considerations",
			})
			return
		}

		slog.Info("webhook.hmac.valid",
			"trace_id", traceID,
		)
		c.Next()
	}
	// @hlv:sec [INPUT_VALIDATION] — body read and validated before processing
	// @hlv log_all_errors
}
