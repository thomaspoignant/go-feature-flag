# Google Cloud Pub/Sub Exporter v2

## What is this exporter

The `pubsubexporterv2` package provides a feature flag evaluation data exporter for Google Cloud Pub/Sub using the modern `cloud.google.com/go/pubsub/v2` library. This exporter sends feature flag evaluation events as messages to a specified Pub/Sub topic in near real-time.

This exporter is a **queue-type** exporter, meaning it sends events individually as soon as they are received, rather than batching them. This enables near real-time processing of feature flag evaluation data.

### Key Benefits of v2

- **Improved Performance**: The v2 library provides better throughput and lower latency
- **Enhanced Features**: Access to the latest Pub/Sub features and optimizations
- **Better Resource Management**: More efficient connection pooling and resource utilization
- **Future-Proof**: Built on the latest Google Cloud client library architecture

## How to configure it

### Basic Configuration

```go
package main

import (
    "log"
    "time"

    "github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporterv2"
    "github.com/thomaspoignant/go-feature-flag/ffclient"
    "github.com/thomaspoignant/go-feature-flag/retriever/httpretriever"
)

func main() {
    config := ffclient.Config{
        PollingInterval: 3 * time.Second,
        Retriever: &httpretriever.Retriever{
            URL: "https://example.com/flags.yaml",
        },
        DataExporter: ffclient.DataExporter{
            FlushInterval:    10 * time.Second,
            MaxEventInMemory: 1000,
            Exporter: &pubsubexporterv2.Exporter{
                ProjectID: "my-gcp-project-id", // required
                Topic:     "feature-flag-events", // required
            },
        },
    }

    err := ffclient.Init(config)
    if err != nil {
        log.Fatal(err)
    }
    defer ffclient.Close()
}
```

### Advanced Configuration

```go
import (
    "runtime"
    "time"

    "cloud.google.com/go/pubsub/v2"
    "google.golang.org/api/option"
)

config := ffclient.Config{
    // ... other config
    DataExporter: ffclient.DataExporter{
        Exporter: &pubsubexporterv2.Exporter{
            ProjectID: "my-gcp-project-id",
            Topic:     "feature-flag-events",
            
            // Optional: Custom client options
            Options: []option.ClientOption{
                option.WithCredentialsFile("/path/to/service-account.json"),
                // or option.WithCredentialsJSON(jsonKey),
            },
            
            // Optional: Publisher settings for batching and performance tuning
            PublishSettings: &pubsub.PublishSettings{
                DelayThreshold:    100 * time.Millisecond,
                CountThreshold:    100,
                ByteThreshold:     1024 * 1024, // 1MB
                NumGoroutines:     runtime.NumCPU(),
                Timeout:           60 * time.Second,
                EnableCompression: true,
                CompressionBytesThreshold: 1024, // 1KB
            },
            
            // Optional: Enable message ordering (requires ordering key in messages)
            EnableMessageOrdering: true,
        },
    },
}
```

### Configuration Options

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ProjectID` | `string` | ✅ | Google Cloud Project ID where the Pub/Sub topic exists |
| `Topic` | `string` | ✅ | Name of the Pub/Sub topic to publish messages to |
| `Options` | `[]option.ClientOption` | ❌ | Google Cloud client options for authentication and configuration |
| `PublishSettings` | `*pubsub.PublishSettings` | ❌ | Publisher settings for batching, compression, and performance tuning |
| `EnableMessageOrdering` | `bool` | ❌ | Enable ordered message delivery (default: false) |

### Authentication

The exporter supports multiple authentication methods:

1. **Default Application Credentials** (recommended for production):
   ```go
   // No Options needed - uses environment or metadata service
   ```

2. **Service Account Key File**:
   ```go
   Options: []option.ClientOption{
       option.WithCredentialsFile("/path/to/service-account.json"),
   }
   ```

3. **Service Account Key JSON**:
   ```go
   Options: []option.ClientOption{
       option.WithCredentialsJSON(jsonKey),
   }
   ```

4. **API Key** (for public topics):
   ```go
   Options: []option.ClientOption{
       option.WithAPIKey("your-api-key"),
   }
   ```

## Migration path for users using pubsubexporter

If you're currently using the legacy `pubsubexporter` (v1), migrating to `pubsubexporterv2` is straightforward:

### Step 1: Update Import

```go
// Before (v1 - deprecated)
import "github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporter"

// After (v2 - recommended)
import "github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporterv2"
```

### Step 2: Update Struct Reference

```go
// Before (v1)
exporter := &pubsubexporter.Exporter{
    ProjectID: "my-project",
    Topic:     "my-topic",
    // ... other fields
}

// After (v2)
exporter := &pubsubexporterv2.Exporter{
    ProjectID: "my-project", 
    Topic:     "my-topic",
    // ... same fields, same configuration
}
```

### Step 3: Update Dependencies (if needed)

The v2 exporter uses `cloud.google.com/go/pubsub/v2`, but this is automatically handled by the Go module system. No manual dependency updates are required.

### Complete Migration Example

```go
// Before (v1 - deprecated)
package main

import (
    "github.com/thomaspoignant/go-feature-flag/ffclient"
    "github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporter"
)

func main() {
    config := ffclient.Config{
        DataExporter: ffclient.DataExporter{
            Exporter: &pubsubexporter.Exporter{
                ProjectID: "my-project",
                Topic:     "feature-events",
            },
        },
    }
    // ...
}

// After (v2 - recommended)
package main

import (
    "github.com/thomaspoignant/go-feature-flag/ffclient"
    "github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporterv2"
)

func main() {
    config := ffclient.Config{
        DataExporter: ffclient.DataExporter{
            Exporter: &pubsubexporterv2.Exporter{
                ProjectID: "my-project",
                Topic:     "feature-events",
            },
        },
    }
    // ...
}
```

### Why Migrate?

- **Performance**: v2 provides significantly better performance and lower latency
- **Features**: Access to newer Pub/Sub features like compression and improved batching
- **Support**: v1 is deprecated and will be removed in a future version
- **Stability**: v2 is built on the latest Google Cloud client architecture
- **Compatibility**: Drop-in replacement with the same configuration options

### Backward Compatibility

The legacy `pubsubexporter` will continue to work but is marked as deprecated. We recommend migrating to `pubsubexporterv2` at your earliest convenience to take advantage of the improvements and ensure future compatibility.

### Need Help?

If you encounter any issues during migration, please:

1. Check that your topic exists and you have proper permissions
2. Verify your authentication configuration
3. Review the [Pub/Sub v2 documentation](https://pkg.go.dev/cloud.google.com/go/pubsub/v2)
4. Open an issue in the [go-feature-flag repository](https://github.com/thomaspoignant/go-feature-flag/issues)