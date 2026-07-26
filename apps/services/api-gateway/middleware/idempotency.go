package middleware

// Idempotency middleware for REST write endpoints.
// Implements REQ-FUN.API.write-idempotency (P0, APPROVED).
//
// Usage:
//
//	r.POST("/groups", idempotencyMiddleware, handler)
//
// Idempotency-Key header must be:
//   - Non-empty alphanumeric + hyphens, max 128 chars
//   - Same key + same body: returns cached response (200/201)
//   - Same key + different body: returns 409 IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_PAYLOAD
//   - New key: executes handler, caches response for 24h
//   - Missing key on critical writes (membership): returns 400 INVALID_IDEMPOTENCY_KEY

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

var (
	idempotencyReplayTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "idempotency_replay_total",
		Help: "Total number of idempotent request replays",
	})
	idempotencyConflictTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "idempotency_conflict_total",
		Help: "Total number of idempotency key conflicts (409)",
	})
)

const (
	idemKeyMaxLen = 128
	idemTTL       = 24 * time.Hour
	idemKeyPrefix = "idem:"
)

// CachedResponse stores the idempotent response to replay.
type CachedResponse struct {
	Status     int               `json:"status"`
	Headers    map[string]string `json:"headers,omitempty"`
	BodyHash   string            `json:"body_hash"`
	Body       json.RawMessage   `json:"body,omitempty"`
	ResourceID string            `json:"resource_id,omitempty"`
	Location   string            `json:"location,omitempty"`
}

// IdempotencyStore defines the storage backend for idempotency keys.
type IdempotencyStore interface {
	Get(ctx context.Context, key string) (*CachedResponse, error)
	Set(ctx context.Context, key string, resp *CachedResponse) error
	Delete(ctx context.Context, key string) error
}

// ============================================================================
// In-Memory Store (for testing and fallback)
// ============================================================================

// MemIdempotencyStore is an in-memory implementation of IdempotencyStore.
type MemIdempotencyStore struct {
	mu   sync.RWMutex
	data map[string]*memEntry
}

type memEntry struct {
	resp      *CachedResponse
	expiresAt time.Time
}

func NewMemIdempotencyStore() *MemIdempotencyStore {
	return &MemIdempotencyStore{
		data: make(map[string]*memEntry),
	}
}

func (s *MemIdempotencyStore) Get(ctx context.Context, key string) (*CachedResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.data[key]
	if !ok {
		return nil, nil
	}
	if time.Now().After(e.expiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		s.mu.RLock()
		return nil, nil
	}
	return e.resp, nil
}

func (s *MemIdempotencyStore) Set(ctx context.Context, key string, resp *CachedResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &memEntry{
		resp:      resp,
		expiresAt: time.Now().Add(idemTTL),
	}
	return nil
}

func (s *MemIdempotencyStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// ============================================================================
// Redis Store
// ============================================================================

// RedisIdempotencyStore is a Redis-backed implementation of IdempotencyStore.
type RedisIdempotencyStore struct {
	client *redis.Client
}

func NewRedisIdempotencyStore(redisURL string) (*RedisIdempotencyStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}
	client := redis.NewClient(opts)
	return &RedisIdempotencyStore{client: client}, nil
}

func (s *RedisIdempotencyStore) Get(ctx context.Context, key string) (*CachedResponse, error) {
	data, err := s.client.Get(ctx, idemKeyPrefix+key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var resp CachedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *RedisIdempotencyStore) Set(ctx context.Context, key string, resp *CachedResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, idemKeyPrefix+key, data, idemTTL).Err()
}

func (s *RedisIdempotencyStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, idemKeyPrefix+key).Err()
}

// ============================================================================
// Middleware
// ============================================================================

// IdempotencyConfig configures the Idempotency middleware.
type IdempotencyConfig struct {
	Store         IdempotencyStore
	CriticalPaths []string // Paths that require Idempotency-Key (returns 400 if missing)
	SkippedPaths  []string // Paths that skip idempotency check
}

// Idempotency returns a Gin middleware that enforces idempotency for write endpoints.
//
//nolint:gocyclo
func Idempotency(cfg *IdempotencyConfig) gin.HandlerFunc {
	store := cfg.Store
	if store == nil {
		store = NewMemIdempotencyStore()
	}

	return func(c *gin.Context) {
		// Skip GET/OPTIONS/HEAD requests
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions || c.Request.Method == http.MethodHead {
			c.Next()
			return
		}

		// Check if path should be skipped
		for _, skipped := range cfg.SkippedPaths {
			if strings.HasPrefix(c.Request.URL.Path, skipped) {
				c.Next()
				return
			}
		}

		key := c.GetHeader("Idempotency-Key")

		// Check if this path requires Idempotency-Key
		isCritical := false
		for _, cp := range cfg.CriticalPaths {
			if strings.HasPrefix(c.Request.URL.Path, cp) {
				isCritical = true
				break
			}
		}

		if key == "" {
			if isCritical {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"code":    "INVALID_IDEMPOTENCY_KEY",
						"message": "Idempotency-Key header is required for this endpoint",
					},
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Validate format
		if !isValidIdemKey(key) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_IDEMPOTENCY_KEY",
					"message": "Idempotency-Key must be alphanumeric with hyphens, max 128 chars",
				},
			})
			c.Abort()
			return
		}

		// Read request body for hash
		bodyBytes, err := c.GetRawData()
		if err != nil {
			slog.Error("idempotency.read_body_failed", "error", err)
			c.Next()
			return
		}
		bodyHash := sha256Hex(bodyBytes)

		// Check for existing key
		ctx := c.Request.Context()
		cached, err := store.Get(ctx, key)
		if err != nil {
			// Degraded mode: Redis unavailable, pass through
			slog.Warn("idempotency.store_read_failed", "key", key, "error", err)
			c.Next()
			return
		}

		if cached != nil {
			if cached.BodyHash == bodyHash {
				// Same key + same body: replay cached response
				idempotencyReplayTotal.Inc()
				for k, v := range cached.Headers {
					c.Header(k, v)
				}
				if cached.Location != "" {
					c.Header("Location", cached.Location)
				}
				c.JSON(cached.Status, cached.Body)
				c.Abort()
				return
			}

			// Same key + different body: conflict
			idempotencyConflictTotal.Inc()
			slog.Warn("idempotency.conflict",
				"key", key,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_PAYLOAD",
					"message": "Idempotency-Key was already used with a different request body",
				},
			})
			c.Abort()
			return
		}

		// New key: wrap the response writer to capture the response
		wrapped := &responseCapture{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = wrapped
		c.Next()

		// Only cache successful responses (2xx)
		if wrapped.status >= 200 && wrapped.status < 300 {
			cachedResp := &CachedResponse{
				Status:   wrapped.status,
				BodyHash: bodyHash,
				Body:     wrapped.body,
			}
			// Extract Location header if present
			if loc := c.Writer.Header().Get("Location"); loc != "" {
				cachedResp.Location = loc
			}
			// Store response headers
			cachedResp.Headers = make(map[string]string)
			for k, v := range c.Writer.Header() {
				if len(v) > 0 {
					cachedResp.Headers[k] = v[0]
				}
			}

			if err := store.Set(ctx, key, cachedResp); err != nil {
				slog.Warn("idempotency.store_write_failed", "key", key, "error", err)
			}
		}
	}
}

// responseCapture wraps gin.ResponseWriter to capture the status code and body.
type responseCapture struct {
	gin.ResponseWriter
	status int
	body   json.RawMessage
}

func (w *responseCapture) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseCapture) Write(data []byte) (int, error) {
	w.body = data
	return w.ResponseWriter.Write(data)
}

// ============================================================================
// Helpers
// ============================================================================

func isValidIdemKey(key string) bool {
	if len(key) == 0 || len(key) > idemKeyMaxLen {
		return false
	}
	for _, c := range key {
		if !isAlphanumericOrHyphen(c) {
			return false
		}
	}
	return true
}

func isAlphanumericOrHyphen(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-'
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
