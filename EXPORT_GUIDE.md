# How to Export SentryKit as Standalone Module

SentryKit is designed to be a reusable, standalone package that can be extracted and used in other projects.

## Option 1: Use as Internal Package (Current Setup)

Currently, SentryKit is part of `alurkerja/probis_service` and can be imported as:

```go
import "alurkerja/probis_service/sentrykit"
```

## Option 2: Extract as Separate Go Module

### Step 1: Copy SentryKit Folder

Copy the entire `sentrykit` folder to a new location:

```bash
cp -r sentrykit /path/to/your-sentrykit-repo
cd /path/to/your-sentrykit-repo
```

### Step 2: Initialize Go Module

```bash
# Initialize with your module path
go mod init github.com/purwadarozatun/go-sentry-fiber-3

# Or for internal use
go mod init company.com/internal/sentrykit
```

### Step 3: Update Dependencies

```bash
go mod tidy
```

This will create a `go.mod` file with required dependencies:

```go
module github.com/purwadarozatun/go-sentry-fiber-3

go 1.21

require (
	github.com/getsentry/sentry-go v0.36.0
	github.com/gofiber/fiber/v3 v3.0.0-beta.3
)
```

### Step 4: Commit to Git Repository

```bash
git init
git add .
git commit -m "Initial commit: SentryKit standalone module"
git remote add origin https://github.com/purwadarozatun/go-sentry-fiber-3.git
git push -u origin main
```

### Step 5: Tag a Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

### Step 6: Use in Other Projects

Now you can use SentryKit in any Go project:

```bash
go get github.com/purwadarozatun/go-sentry-fiber-3@v1.0.0
```

```go
import "github.com/purwadarozatun/go-sentry-fiber-3"

func main() {
    sentrykit.Init(sentrykit.Config{
        DSN: "your-dsn",
        Environment: "production",
    })
    defer sentrykit.Close()
    
    // Use with Fiber
    app := fiber.New()
    app.Use(sentrykit.New())
}
```

## Option 3: Use as Git Submodule

### In the Source Project (probis-service)

```bash
# Remove sentrykit from tracking but keep files
git rm --cached -r sentrykit
git commit -m "Prepare sentrykit for submodule"

# Create separate repo for sentrykit
cd sentrykit
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/purwadarozatun/go-sentry-fiber-3.git
git push -u origin main
```

### In Projects That Want to Use It

```bash
# Add as submodule
git submodule add https://github.com/purwadarozatun/go-sentry-fiber-3.git sentrykit

# Update go.mod to use local path
go mod edit -replace github.com/purwadarozatun/go-sentry-fiber-3=./sentrykit
```

## Option 4: Private Go Module (GitLab/GitHub Enterprise)

### Setup Private Module

1. Push to private repository:

```bash
git remote add origin https://gitlab.company.com/libs/sentrykit.git
git push -u origin main
git tag v1.0.0
git push origin v1.0.0
```

2. Configure Go to access private repo:

```bash
# For GitLab
go env -w GOPRIVATE=gitlab.company.com/libs/*

# For GitHub Enterprise  
go env -w GOPRIVATE=github.company.com/libs/*
```

3. Use in projects:

```bash
go get gitlab.company.com/libs/sentrykit@v1.0.0
```

## File Structure for Standalone Module

When extracting, the structure should be:

```
sentrykit/
├── README.md              # Full documentation
├── LICENSE               # Your license file
├── go.mod                # Module definition
├── client.go             # Core Sentry functions
├── middleware.go         # Fiber middleware
├── examples/             # Example usage
│   └── fiber/
│       └── main.go
└── .gitignore
```

## Creating .gitignore

```bash
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~
```

## Versioning Strategy

Follow Semantic Versioning (SemVer):

- **v1.0.0** - Initial stable release
- **v1.0.1** - Bug fixes
- **v1.1.0** - New features (backward compatible)
- **v2.0.0** - Breaking changes

Example:
```bash
git tag v1.0.0
git push origin v1.0.0

# After adding new features
git tag v1.1.0
git push origin v1.1.0
```

## Publishing to Public Registry

### GitHub

1. Push to GitHub public repository
2. Tag releases
3. Users can install with:
   ```bash
   go get github.com/purwadarozatun/go-sentry-fiber-3@latest
   ```

### pkg.go.dev

Once published on GitHub, it will automatically appear on https://pkg.go.dev/github.com/purwadarozatun/go-sentry-fiber-3

## Testing the Standalone Module

Before publishing:

```bash
# In sentrykit directory
go test ./...

# Check module
go mod verify

# Try building examples
cd examples/fiber
go build .
```

## Usage After Export

### Basic Import

```go
package main

import (
    "log"
    "github.com/purwadarozatun/go-sentry-fiber-3"
    "github.com/gofiber/fiber/v3"
)

func main() {
    // Initialize
    if err := sentrykit.Init(sentrykit.Config{
        DSN: "your-sentry-dsn",
        Environment: "production",
    }); err != nil {
        log.Fatal(err)
    }
    defer sentrykit.Close()

    // Use with Fiber
    app := fiber.New()
    app.Use(sentrykit.New())
    
    app.Get("/", func(c fiber.Ctx) error {
        sentrykit.AddBreadcrumbFromContext(c, "Home page", "http", nil)
        return c.SendString("Hello!")
    })
    
    log.Fatal(app.Listen(":3000"))
}
```

## Maintenance

### Updating Dependencies

```bash
go get -u github.com/getsentry/sentry-go
go get -u github.com/gofiber/fiber/v3
go mod tidy
```

### Release New Version

```bash
git add .
git commit -m "Update dependencies"
git tag v1.0.1
git push origin main
git push origin v1.0.1
```

## Benefits of Standalone Module

1. ✅ **Reusable** - Use across multiple projects
2. ✅ **Versioned** - Stable releases with semantic versioning
3. ✅ **Maintainable** - Separate repository, easier to manage
4. ✅ **Testable** - Independent test suite
5. ✅ **Documented** - Clear API documentation
6. ✅ **Portable** - No dependencies on probis-service
7. ✅ **Shareable** - Easy to share with team or community

## Current Usage in probis-service

The sentrykit package is currently used in probis-service as an internal package. It's fully functional and ready to be extracted whenever needed.

```go
// main.go
import "alurkerja/probis_service/sentrykit"

sentrykit.Init(sentrykit.Config{
    DSN: config.Config.SentryDSN,
    Environment: config.Config.SentryEnvironment,
})
defer sentrykit.Close()

app.Use(sentrykit.New())
```

## Next Steps

1. ✅ Package is ready and tested
2. ⏳ Choose export option (1-4 above)
3. ⏳ Create standalone repository
4. ⏳ Tag first release
5. ⏳ Update probis-service to use external module (optional)
