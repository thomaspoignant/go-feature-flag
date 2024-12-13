---
sidebar_position: 8
---

# Azure Blob Storage
This retriever is used to retrieve data from a Container on Azure Blob Storage.

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

## Example
```go showLineNumbers
awsConfig, _ := config.LoadDefaultConfig(context.Background())
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &azblobretriever.Retriever{
        Container:    "your-container",
        AccountName:  "your-account-name",
        AccountKey:   "your-account-key",
        Object:       "your_location/feature-flags.json",
    },
})
defer ffclient.Close()
```
## Configuration fields
To configure your S3 file location:

| Field             | Description                                     |
|-------------------|-------------------------------------------------|
| **`Container`**   | Name of the Azure Blob Storage container        |
| **`Object`**      | Name of the feature flag file in the container  |
| **`AccountName`** | Azure Storage Account Name                      |
| **`AccountKey`**  | (Optional) Storage Account Key                  |
| **`ServiceURL`**  | (Optional) Custom service URL                   |

## Authentication Methods

### 1. Shared Key Authentication

```go
retriever := &azblobretriever.Retriever{
    Container:    "your-container",
    AccountName:  "your-account-name",
    AccountKey:   "your-account-key",
    Object:       "feature-flags.json",
}
```

### 2. Microsoft Entra ID (Recommended)

```go
retriever := &azblobretriever.Retriever{
    Container:    "your-container",
    AccountName:  "your-account-name",
    Object:       "feature-flags.json",
}
```


