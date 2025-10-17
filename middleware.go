package sentrykit

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
)

// MiddlewareConfig holds configuration for Fiber middleware
type MiddlewareConfig struct {
	// Repanic configures whether Sentry should repanic after recovery
	Repanic bool
	
	// WaitForDelivery configures whether to block/wait until events are sent
	WaitForDelivery bool
	
	// Timeout for event delivery
	Timeout time.Duration
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		Repanic:         false,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	}
}

// New creates a new Sentry middleware for Fiber
func New(config ...MiddlewareConfig) fiber.Handler {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c fiber.Ctx) error {
		// Create a new hub for this request
		hub := sentry.CurrentHub().Clone()

		// Add request context
		hub.Scope().SetContext("request", map[string]interface{}{
			"url":          c.OriginalURL(),
			"method":       c.Method(),
			"query_string": string(c.Request().URI().QueryString()),
			"headers":      extractHeaders(c),
			"ip":           c.IP(),
			"user_agent":   c.Get("User-Agent"),
		})

		// Add custom tags
		hub.Scope().SetTag("path", c.Path())
		hub.Scope().SetTag("method", c.Method())

		// Extract and set user info if available
		if userID := c.Locals("user_id"); userID != nil {
			hub.Scope().SetUser(sentry.User{
				ID: fmt.Sprintf("%v", userID),
			})
		}

		// Extract tenant ID from params if available
		if tenantID := c.Params("tenantId"); tenantID != "" {
			hub.Scope().SetTag("tenant_id", tenantID)
		}

		// Store hub in context for later use
		c.Locals("sentry_hub", hub)

		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				hub.Recover(err)
				
				if cfg.WaitForDelivery {
					hub.Flush(cfg.Timeout)
				}
				
				if cfg.Repanic {
					panic(err)
				}
			}
		}()

		// Process request
		err := c.Next()

		// Capture errors (5xx only)
		if err != nil {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Only capture server errors (5xx)
			if code >= 500 {
				hub.CaptureException(err)
				
				// Add error context
				hub.Scope().SetContext("error_details", map[string]interface{}{
					"error":      err.Error(),
					"path":       c.Path(),
					"method":     c.Method(),
					"status":     code,
					"ip":         c.IP(),
					"user_agent": c.Get("User-Agent"),
				})
			}
		}

		// Flush events if configured
		if cfg.WaitForDelivery {
			hub.Flush(cfg.Timeout)
		}

		return err
	}
}

// extractHeaders extracts HTTP headers and filters sensitive ones
func extractHeaders(c fiber.Ctx) map[string]string {
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		keyStr := string(key)
		// Skip sensitive headers
		if keyStr != "Authorization" && keyStr != "Cookie" && keyStr != "X-Api-Key" {
			headers[keyStr] = string(value)
		}
	})
	return headers
}

// GetHubFromContext retrieves the Sentry hub from Fiber context
func GetHubFromContext(c fiber.Ctx) *sentry.Hub {
	if hub, ok := c.Locals("sentry_hub").(*sentry.Hub); ok {
		return hub
	}
	return sentry.CurrentHub()
}

// CaptureExceptionFromContext captures an exception using the hub from context
func CaptureExceptionFromContext(c fiber.Ctx, err error) *sentry.EventID {
	hub := GetHubFromContext(c)
	return hub.CaptureException(err)
}

// CaptureMessageFromContext captures a message using the hub from context
func CaptureMessageFromContext(c fiber.Ctx, message string, level sentry.Level) *sentry.EventID {
	hub := GetHubFromContext(c)
	hub.Scope().SetLevel(level)
	return hub.CaptureMessage(message)
}

// AddBreadcrumbFromContext adds a breadcrumb using the hub from context
func AddBreadcrumbFromContext(c fiber.Ctx, message, category string, data map[string]interface{}) {
	hub := GetHubFromContext(c)
	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Message:  message,
		Category: category,
		Data:     data,
		Level:    sentry.LevelInfo,
	}, nil)
}

// SetUserFromContext sets user information using the hub from context
func SetUserFromContext(c fiber.Ctx, userID, email, username string) {
	hub := GetHubFromContext(c)
	hub.Scope().SetUser(sentry.User{
		ID:       userID,
		Email:    email,
		Username: username,
	})
}

// SetTagFromContext sets a tag using the hub from context
func SetTagFromContext(c fiber.Ctx, key, value string) {
	hub := GetHubFromContext(c)
	hub.Scope().SetTag(key, value)
}

// SetContextFromContext sets context data using the hub from context
func SetContextFromContext(c fiber.Ctx, key string, data map[string]interface{}) {
	hub := GetHubFromContext(c)
	hub.Scope().SetContext(key, data)
}
