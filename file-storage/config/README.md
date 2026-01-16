# Filesystems Configuration Package

Complete configuration system for file storage with support for multiple backends.

## Overview

The `config` package provides a flexible, validated configuration system for:
- **Local disk storage** (development, default)
- **AWS S3** (production)
- **Google Cloud Storage** (enterprise)
- **Custom drivers** (extend as needed)

## Quick Start

### 1. Load from Environment

```go
package main

import "bisma-api/modules/file-storage/config"

func init() {
    cfg, err := config.LoadFromEnv()
    if err != nil {
        panic(err)
    }
    
    if err := cfg.Validate(); err != nil {
        panic(err)
    }
}
```

### 2. Programmatic Configuration

```go
cfg := config.New().
    WithLocalDriver("storage/app/uploads", 0, nil).
    SetDefault("local")

if err := cfg.Validate(); err != nil {
    panic(err)
}
```

### 3. Multi-Driver Setup

```go
cfg := config.New().
    WithLocalDriver("storage/app/uploads", 0, nil).
    WithS3Driver("ap-southeast-1", "my-bucket", "", 0, nil).
    WithGCSDriver("my-project", "my-bucket", "", "/path/to/sa.json", 0, nil).
    SetDefault("local")

if err := cfg.Validate(); err != nil {
    panic(err)
}
```

## Configuration Types

### LocalConfig

Configuration for local disk storage.

```go
type LocalConfig struct {
    Path         string   // Directory for uploads
    MaxSize      int64    // Max file size (bytes)
    AllowedTypes []string // MIME type whitelist
}
```

**Environment Variables**:
```env
FILESYSTEM_LOCAL_PATH=storage/app/uploads
FILESYSTEM_MAX_SIZE=10485760
```

### S3Config

Configuration for AWS S3.

```go
type S3Config struct {
    Region       string   // AWS region
    Bucket       string   // S3 bucket
    BaseURL      string   // Download URL
    AccessKey    string   // AWS access key
    SecretKey    string   // AWS secret key
    MaxSize      int64    // Max file size
    AllowedTypes []string // MIME types
    ACL          string   // Object ACL
    StorageClass string   // Storage class
}
```

**Environment Variables**:
```env
AWS_REGION=ap-southeast-1
S3_BUCKET=my-bucket
S3_BASE_URL=https://my-bucket.s3.amazonaws.com
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
S3_ACL=private
S3_STORAGE_CLASS=STANDARD
```

### GCSConfig

Configuration for Google Cloud Storage.

```go
type GCSConfig struct {
    Project         string   // GCP project ID
    Bucket          string   // GCS bucket
    BaseURL         string   // Download URL
    CredentialsFile string   // Service account JSON
    MaxSize         int64    // Max file size
    AllowedTypes    []string // MIME types
}
```

**Environment Variables**:
```env
GCP_PROJECT=my-project
GCS_BUCKET=my-bucket
GCS_BASE_URL=https://storage.googleapis.com/my-bucket
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

## API Reference

### New()

Create a new Config instance.

```go
cfg := config.New()
```

### WithLocalDriver()

Add local storage driver.

```go
cfg.WithLocalDriver(
    "storage/app/uploads",  // path
    0,                      // max size (0 = default)
    nil,                    // allowed types (nil = default)
)
```

### WithS3Driver()

Add S3 driver.

```go
cfg.WithS3Driver(
    "ap-southeast-1",       // region
    "my-bucket",            // bucket
    "https://...",          // base URL
    0,                      // max size (0 = default)
    nil,                    // allowed types (nil = default)
)
```

### WithGCSDriver()

Add GCS driver.

```go
cfg.WithGCSDriver(
    "my-project",           // project
    "my-bucket",            // bucket
    "https://...",          // base URL
    "/path/to/sa.json",     // credentials file
    0,                      // max size (0 = default)
    nil,                    // allowed types (nil = default)
)
```

### SetDefault()

Set the default driver.

```go
cfg.SetDefault("local")
```

### GetDriver()

Get driver configuration by name.

```go
driver, err := cfg.GetDriver("s3")
if err != nil {
    // Handle error
}
```

### GetDefaultDriver()

Get the default driver configuration.

```go
defaultDriver, err := cfg.GetDefaultDriver()
if err != nil {
    // Handle error
}
```

### Validate()

Validate entire configuration.

```go
if err := cfg.Validate(); err != nil {
    panic(err)
}
```

### String()

Get human-readable configuration.

```go
println(cfg.String())
// Output:
// Filesystems Configuration:
// Default Driver: local
// Configured Drivers:
//   - local (local)
//   - s3 (s3)
```

## Usage Patterns

### Pattern 1: Environment-Based

```go
import (
    "os"
    "bisma-api/modules/file-storage/config"
    "bisma-api/modules/file-storage/factory"
    "bisma-api/modules/file-storage/providers"
)

func init() {
    cfg, err := config.LoadFromEnv()
    if err != nil {
        panic(err)
    }
    
    if err := cfg.Validate(); err != nil {
        panic(err)
    }
    
    // Initialize factory with configured drivers
    providerMap := make(map[string]filestorage.FileStorageProvider)
    
    for name := range cfg.Drivers {
        switch name {
        case "local":
            localCfg := cfg.Drivers["local"].(*config.LocalConfig)
            providerMap["local"] = providers.NewLocalStorage(localCfg)
        case "s3":
            // Initialize S3 provider with config
        case "gcs":
            // Initialize GCS provider with config
        }
    }
    
    factory.Initialize(cfg.Default, providerMap)
}
```

### Pattern 2: Development Multi-Driver

```go
cfg := config.New().
    WithLocalDriver("storage/app/uploads", 0, nil).
    WithS3Driver("ap-southeast-1", "dev-bucket", "", 0, nil).
    SetDefault("local")

// Use local by default, S3 available for testing
```

### Pattern 3: Production with Fallback

```go
cfg, err := config.LoadFromEnv()

// Ensure we have at least local fallback
if _, err := cfg.GetDriver("local"); err != nil {
    cfg.WithLocalDriver("storage/app/uploads", 0, nil)
}

if err := cfg.Validate(); err != nil {
    panic(err)
}
```

### Pattern 4: Custom Validation

```go
cfg, err := config.LoadFromEnv()
if err != nil {
    panic(err)
}

// Add custom validation
if err := cfg.Validate(); err != nil {
    panic(err)
}

// Check default driver is S3 in production
if os.Getenv("ENVIRONMENT") == "production" && cfg.Default != "s3" {
    panic("S3 must be default driver in production")
}
```

## Defaults

### Default Allowed MIME Types

```go
config.DefaultAllowedTypes()
// Returns:
// [
//     "image/jpeg",
//     "image/jpg",
//     "image/png",
//     "image/gif",
//     "application/pdf",
//     "application/msword",
//     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
//     "application/vnd.ms-excel",
//     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
// ]
```

### Default Max Size

```go
config.DefaultMaxSize()
// Returns: 10485760 (10MB)
```

## Validation

Configuration is validated for:

✓ **Presence**
- Default driver is set
- Drivers are configured

✓ **Completeness**
- Required fields present
- Credentials available

✓ **Consistency**
- Default driver exists
- Each driver configuration is valid

**Validation Example**:

```go
cfg := config.New().WithLocalDriver("", 0, nil).SetDefault("local")

err := cfg.Validate()
// Error: "driver local validation failed: local driver: path is required"
```

## Error Handling

Common errors and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| `no default driver configured` | SetDefault() not called | Call SetDefault("local") |
| `default driver X not configured` | Driver not added | Use WithLocalDriver(), etc. |
| `no drivers configured` | No drivers added | Add at least one driver |
| `path is required` | LocalConfig.Path empty | Set valid path |
| `bucket is required` | S3Config.Bucket empty | Set bucket name |
| `credentials_file not found` | GCS file missing | Verify file exists |

## Security Best Practices

### Credentials Management

✓ Use environment variables for sensitive data
✓ Never hardcode credentials
✓ Use service accounts (GCS)
✓ Rotate credentials regularly
✓ Use IAM roles where possible

**Good**:
```go
cfg, err := config.LoadFromEnv() // Reads from .env or env vars
```

**Bad**:
```go
cfg := config.New().
    WithS3Driver("region", "bucket", "", 0, nil)
cfg.Drivers["s3"].(*config.S3Config).AccessKey = "AKIAIOSFODNN7EXAMPLE"
```

### File Permissions

✓ Store GCS credentials with restricted permissions
✓ Use `chmod 600` for credential files
✓ Store files outside web root (local)
✓ Use signed URLs for downloads

### MIME Type Validation

✓ Whitelist allowed types
✓ Validate on server (not client)
✓ Don't trust Content-Type header alone
✓ Validate magic bytes for critical uploads

## Integration with Factory

```go
import (
    "bisma-api/modules/file-storage/config"
    "bisma-api/modules/file-storage/factory"
    "bisma-api/modules/file-storage/providers"
)

func initializeFilesystem() {
    // 1. Load configuration
    cfg, err := config.LoadFromEnv()
    if err != nil {
        panic(err)
    }
    
    // 2. Validate
    if err := cfg.Validate(); err != nil {
        panic(err)
    }
    
    // 3. Create providers from config
    providerMap := map[string]filestorage.FileStorageProvider{
        "local": providers.NewLocalStorage(
            cfg.Drivers["local"].(*config.LocalConfig),
        ),
    }
    
    // 4. Initialize factory
    factory.Initialize(cfg.Default, providerMap)
}
```

## Examples

See `examples.go` for complete working examples:
- Basic local setup
- Environment loading
- Multi-driver configuration
- Custom types/sizes
- Error handling
- Environment-based selection
- Programmatic fallbacks
- And more...

## Testing

### Mock Configuration

```go
func createTestConfig() *config.Config {
    return config.New().
        WithLocalDriver("storage/test", 1024*1024, nil). // 1MB for testing
        SetDefault("local")
}
```

### Configuration Validation

```go
func TestConfig(t *testing.T) {
    cfg := createTestConfig()
    
    if err := cfg.Validate(); err != nil {
        t.Fatalf("Config validation failed: %v", err)
    }
    
    driver, err := cfg.GetDefaultDriver()
    if err != nil {
        t.Fatalf("Get default driver failed: %v", err)
    }
    
    if driver.Driver() != "local" {
        t.Errorf("Expected local driver, got %s", driver.Driver())
    }
}
```

## Summary

The `config` package provides:

✅ Flexible configuration management
✅ Multiple driver support
✅ Environment variable loading
✅ Comprehensive validation
✅ Type-safe API
✅ Builder pattern
✅ Easy integration

Ready for development, staging, and production use!
