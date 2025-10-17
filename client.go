package sentrykit

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

// Config holds Sentry configuration
type Config struct {
	DSN              string  // Sentry DSN from your project settings
	Environment      string  // Environment name (development, staging, production)
	Release          string  // Application release/version (optional)
	TracesSampleRate float64 // Percentage of transactions to sample (0.0 - 1.0)
	Debug            bool    // Enable debug mode
	AttachStacktrace bool    // Attach stack traces to messages
	ServerName       string  // Server/host name (optional)
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Environment:      "development",
		TracesSampleRate: 1.0,
		Debug:            false,
		AttachStacktrace: true,
	}
}

// Init initializes Sentry with the provided configuration
func Init(cfg Config) error {
	if cfg.DSN == "" {
		return fmt.Errorf("sentry DSN is required")
	}

	// Set default environment if not provided
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		TracesSampleRate: cfg.TracesSampleRate,
		Debug:            cfg.Debug,
		AttachStacktrace: cfg.AttachStacktrace,
		ServerName:       cfg.ServerName,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Filter sensitive data if needed
			// This is a hook where you can modify events before sending
			return event
		},
	})

	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	return nil
}

// Close flushes buffered events and closes the Sentry client
func Close() {
	sentry.Flush(2 * time.Second)
}

// CaptureException captures an error and sends it to Sentry
func CaptureException(err error) *sentry.EventID {
	return sentry.CaptureException(err)
}

// CaptureMessage captures a message and sends it to Sentry
func CaptureMessage(message string, level sentry.Level) *sentry.EventID {
	event := sentry.CaptureMessage(message)
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
	})
	return event
}

// RecoverWithSentry recovers from panic and sends to Sentry
func RecoverWithSentry() {
	if err := recover(); err != nil {
		sentry.CurrentHub().Recover(err)
		sentry.Flush(2 * time.Second)
	}
}

// AddBreadcrumb adds a breadcrumb to the current scope
func AddBreadcrumb(message, category string, data map[string]interface{}) {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Message:  message,
		Category: category,
		Data:     data,
		Level:    sentry.LevelInfo,
	})
}

// SetUser sets the user context in Sentry
func SetUser(userID, email, username string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       userID,
			Email:    email,
			Username: username,
		})
	})
}

// SetTag sets a custom tag
func SetTag(key, value string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

// SetContext sets custom context data
func SetContext(key string, data map[string]interface{}) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetContext(key, data)
	})
}

// WithScope runs a function with a custom Sentry scope
func WithScope(f func(scope *sentry.Scope)) {
	sentry.WithScope(f)
}

// ConfigureScope configures the current scope
func ConfigureScope(f func(scope *sentry.Scope)) {
	sentry.ConfigureScope(f)
}
