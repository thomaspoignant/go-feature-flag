# Azure Blob Storage Feature Flag Retriever

This retriever is used to retrieve data from a Container on Azure Blob Storage.

## Installation

```bash
go get github.com/thomaspoignant/go-feature-flag/retriever/azblobstorageretriever
```

## Usage

### Configuration

Create a `Retriever` struct with the following fields:

- `Container`: Name of the Azure Blob Storage container
- `AccountName`: Azure Storage Account Name
- `AccountKey`: (Optional) Storage Account Key
- `ServiceURL`: (Optional) Custom service URL
- `Object`: Name of the feature flag file in the container

### Authentication Methods

#### 1. Shared Key Authentication

```go
retriever := &azblobretriever.Retriever{
    Container:    "your-container",
    AccountName:  "your-account-name",
    AccountKey:   "your-account-key",
    Object:       "feature-flags.json",
}
```

#### 2. Microsoft Entra ID (Recommended)

```go
retriever := &azblobretriever.Retriever{
    Container:    "your-container",
    AccountName:  "your-account-name",
    Object:       "feature-flags.json",
}
```

### Retrieving Feature Flags

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/thomaspoignant/go-feature-flag/retriever/azblobstorageretriever"
)

func main() {
    retriever := &azblobretriever.Retriever{
        Container:   "feature-flags",
        AccountName: "mystorageaccount",
        AccountKey:  "your-account-key",
        Object:      "flags.json",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    err := retriever.Init(ctx, nil)
    defer func() { _ = r.Shutdown(ctx) }()
    if err != nil {
        log.Fatalf("Failed to initialize retriever:", err)
    }

    data, err := retriever.Retrieve(ctx)
    if err != nil {
        log.Fatalf("Failed to retrieve feature flags: %v", err)
    }

    fmt.Println("Retrieved feature flags:")
    fmt.Println(string(data))
}
```

## Key Features

- Supports both shared key and default Azure credential authentication
- Automatic retry mechanism for blob downloads
- Flexible configuration with optional custom service URL

## Error Handling

The `Retrieve` method returns an error if:
- `AccountName` is empty
- `Container` or `Object` is not specified
- There's an issue initializing the Azure client.
- There's a problem downloading or reading the file from Azure Blob Storage.

## Best Practices

- **Security** Never hard-code your `AccountKey` in your source code. Use environment variables or a secure secret management system.
- **Error Handling**: Always check for errors returned by the `Retrieve` method.
- **Context**: Use a context with timeout for better control over the retrieval process.
