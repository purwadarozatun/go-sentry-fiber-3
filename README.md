# SentryKit - Sentry Integration for Go Fiber

A standalone, reusable Sentry integration package for Go Fiber applications with comprehensive error tracking, breadcrumbs, and context management.

## Features

- 🎯 Easy Sentry initialization
- 🔌 Fiber middleware with automatic error capture
- 📊 Breadcrumbs tracking
- 👤 User context management
- 🏷️ Custom tags and contexts
- 🔒 Automatic sensitive data filtering
- ⚡ Per-request Sentry hub isolation
- 🛡️ Panic recovery
- 📝 Structured logging support

## Installation

For external projects using this module:
```bash
go get github.com/purwadarozatun/go-sentry-fiber-3@v1.0.0
```

For development within this module:
```bash
# No installation needed - you're already in the module
# Just use: import "github.com/purwadarozatun/go-sentry-fiber-3"
```

This module requires:
- Go 1.21 or higher  
- Fiber v3.0.0-beta.3 or higher
- Sentry Go SDK v0.36.0 or higher

## Quick Start

### 1. Initialize Sentry

```go
package main

import (
    "log"
    sentrykit "github.com/purwadarozatun/go-sentry-fiber-3"
    "github.com/gofiber/fiber/v3"
)

func main() {
    // Initialize Sentry
    err := sentrykit.Init(sentrykit.Config{
        DSN:              "https://your-dsn@sentry.io/project-id",
        Environment:      "production",
        Release:          "v1.0.0",
        TracesSampleRate: 0.2, // 20% of transactions
        Debug:            false,
        AttachStacktrace: true,
    })
    if err != nil {
        log.Fatalf("Sentry initialization failed: %v", err)
    }
    defer sentrykit.Close()

    // Create Fiber app
    app := fiber.New()
    
    // Add Sentry middleware
    app.Use(sentrykit.New())
    
    // Your routes...
    app.Get("/", handler)
    
    app.Listen(":3000")
}
```

### 2. Use in Handlers

```go
func handler(c fiber.Ctx) error {
    // Add breadcrumb
    sentrykit.AddBreadcrumbFromContext(c, "Processing request", "http", map[string]interface{}{
        "endpoint": "/api/users",
    })
    
    // Set user context
    sentrykit.SetUserFromContext(c, "user-123", "user@example.com", "john_doe")
    
    // Add custom tags
    sentrykit.SetTagFromContext(c, "feature", "user_management")
    
    // Add structured context
    sentrykit.SetContextFromContext(c, "business_data", map[string]interface{}{
        "account_id": "ACC-123",
        "plan": "premium",
    })
    
    // Your business logic...
    if err := doSomething(); err != nil {
        // Manually capture error
        sentrykit.CaptureExceptionFromContext(c, err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    
    return c.JSON(fiber.Map{"status": "ok"})
}
```

## API Reference

### Initialization

#### `Init(config Config) error`

Initialize Sentry with configuration.

```go
type Config struct {
    DSN              string  // Required: Your Sentry DSN
    Environment      string  // Environment name (development, staging, production)
    Release          string  // Application version/release
    TracesSampleRate float64 // Sample rate for transactions (0.0 - 1.0)
    Debug            bool    // Enable debug logging
    AttachStacktrace bool    // Attach stack traces to messages
    ServerName       string  // Server identifier
}
```

#### `DefaultConfig() Config`

Returns default configuration values.

#### `Close()`

Flushes buffered events and closes Sentry client. Should be called on shutdown.

### Middleware

#### `New(config ...MiddlewareConfig) fiber.Handler`

Creates Fiber middleware with optional configuration.

```go
type MiddlewareConfig struct {
    Repanic         bool          // Repanic after recovery (default: false)
    WaitForDelivery bool          // Wait for event delivery (default: false)
    Timeout         time.Duration // Flush timeout (default: 2s)
}
```

**Usage:**

```go
// With default config
app.Use(sentrykit.New())

// With custom config
app.Use(sentrykit.New(sentrykit.MiddlewareConfig{
    Repanic:         false,
    WaitForDelivery: true,
    Timeout:         5 * time.Second,
}))
```

### Global Functions

#### `CaptureException(err error) *sentry.EventID`

Capture an error globally (without request context).

#### `CaptureMessage(message string, level sentry.Level) *sentry.EventID`

Capture a message globally.

```go
sentrykit.CaptureMessage("Something important happened", sentry.LevelWarning)
```

#### `RecoverWithSentry()`

Recover from panic and send to Sentry. Use in defer:

```go
defer sentrykit.RecoverWithSentry()
```

#### `AddBreadcrumb(message, category string, data map[string]interface{})`

Add a breadcrumb globally.

#### `SetUser(userID, email, username string)`

Set user context globally.

#### `SetTag(key, value string)`

Set a custom tag globally.

#### `SetContext(key string, data map[string]interface{})`

Set custom context data globally.

### Context-Aware Functions (for use within Fiber handlers)

These functions use the Sentry hub from the request context:

#### `CaptureExceptionFromContext(c fiber.Ctx, err error) *sentry.EventID`

Capture an error with request context.

#### `CaptureMessageFromContext(c fiber.Ctx, message string, level sentry.Level) *sentry.EventID`

Capture a message with request context.

#### `AddBreadcrumbFromContext(c fiber.Ctx, message, category string, data map[string]interface{})`

Add a breadcrumb with request context.

#### `SetUserFromContext(c fiber.Ctx, userID, email, username string)`

Set user information with request context.

#### `SetTagFromContext(c fiber.Ctx, key, value string)`

Set a tag with request context.

#### `SetContextFromContext(c fiber.Ctx, key string, data map[string]interface{})`

Set structured context data with request context.

#### `GetHubFromContext(c fiber.Ctx) *sentry.Hub`

Get the Sentry hub from Fiber context.

## Examples

### Complete Example with Error Handling

```go
package main

import (
    "log"
    "yourapp/sentrykit"
    "github.com/getsentry/sentry-go"
    "github.com/gofiber/fiber/v3"
)

func main() {
    // Initialize Sentry
    if err := sentrykit.Init(sentrykit.Config{
        DSN:              "https://your-dsn@sentry.io/project",
        Environment:      "production",
        TracesSampleRate: 0.2,
    }); err != nil {
        log.Fatal(err)
    }
    defer sentrykit.Close()

    app := fiber.New(fiber.Config{
        ErrorHandler: customErrorHandler,
    })

    // Global middleware
    app.Use(sentrykit.New())

    // Routes
    app.Post("/api/orders", createOrder)
    app.Get("/api/orders/:id", getOrder)

    log.Fatal(app.Listen(":3000"))
}

func customErrorHandler(c fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    if e, ok := err.(*fiber.Error); ok {
        code = e.Code
    }

    // Capture 5xx errors
    if code >= 500 {
        sentrykit.CaptureExceptionFromContext(c, err)
    }

    return c.Status(code).JSON(fiber.Map{
        "error": err.Error(),
    })
}

func createOrder(c fiber.Ctx) error {
    // Add breadcrumb
    sentrykit.AddBreadcrumbFromContext(c, "Creating order", "order", nil)

    var order Order
    if err := c.BodyParser(&order); err != nil {
        return fiber.NewError(400, "Invalid request")
    }

    // Set user context (assuming you have auth middleware)
    userID := c.Locals("user_id").(string)
    sentrykit.SetUserFromContext(c, userID, "", "")

    // Add business context
    sentrykit.SetContextFromContext(c, "order", map[string]interface{}{
        "items_count": len(order.Items),
        "total":       order.Total,
    })

    // Business logic
    if err := processOrder(order); err != nil {
        sentrykit.CaptureExceptionFromContext(c, err)
        return fiber.NewError(500, "Failed to process order")
    }

    return c.JSON(order)
}

func getOrder(c fiber.Ctx) error {
    orderID := c.Params("id")
    
    sentrykit.AddBreadcrumbFromContext(c, "Fetching order", "database", map[string]interface{}{
        "order_id": orderID,
    })

    // Your logic here...
    return c.JSON(fiber.Map{"id": orderID})
}
```

### Using with Background Workers

```go
func backgroundWorker() {
    defer sentrykit.RecoverWithSentry()

    sentrykit.AddBreadcrumb("Worker started", "worker", nil)
    
    if err := doWork(); err != nil {
        sentrykit.CaptureException(err)
    }
}
```

### Custom Scope

```go
sentrykit.WithScope(func(scope *sentry.Scope) {
    scope.SetTag("worker_id", "worker-1")
    scope.SetLevel(sentry.LevelWarning)
    sentrykit.CaptureMessage("Worker restarted", sentry.LevelWarning)
})
```

## Security

The middleware automatically filters sensitive headers:
- `Authorization`
- `Cookie`
- `X-Api-Key`

To filter additional data, modify the `BeforeSend` hook in `client.go`:

```go
BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    // Filter sensitive fields
    if event.Request != nil {
        // Remove sensitive query params
        delete(event.Request.QueryString, "api_key")
        delete(event.Request.QueryString, "token")
    }
    return event
},
```

## Best Practices

1. **Initialize early**: Call `Init()` at the start of your application
2. **Always defer Close()**: Ensures events are flushed on shutdown
3. **Use context-aware functions**: In Fiber handlers, use `FromContext` variants
4. **Add breadcrumbs**: Help debug by tracking user journey
5. **Set user context**: Identify who experienced the error
6. **Use appropriate sample rates**: Lower rate in production (0.1 - 0.3)
7. **Don't capture 4xx errors**: These are client errors, not bugs
8. **Add business context**: Include relevant business data for debugging

## Performance Considerations

- Set appropriate `TracesSampleRate` (0.1-0.3 for production)
- Use `WaitForDelivery: false` for better performance
- Breadcrumbs are stored in memory (don't add too many)
- Events are sent asynchronously by default

## Testing

```go
// For testing, use a test DSN or disable Sentry
if os.Getenv("ENVIRONMENT") != "test" {
    sentrykit.Init(sentrykit.Config{
        DSN: os.Getenv("SENTRY_DSN"),
    })
}
```

## License

MIT License - Feel free to use in your projects

## Contributing

Contributions are welcome! This package is designed to be standalone and reusable.

## Support

For issues or questions:
- Check Sentry documentation: https://docs.sentry.io/platforms/go/
- Open an issue in your repository
