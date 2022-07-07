# Google Cloud Storage

The [**Google Cloud Storage Retriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever/#Retriever) 
will use the [google-cloud-storage package](https://pkg.go.dev/cloud.google.com/go/storage)
and [google-api-options package](https://pkg.go.dev/google.golang.org/api/option) to access your flag in Google Cloud
Storage.

## Example

```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &gcstorageretriever.Retriever{
	    Options: []option.ClientOption{option.WithoutAuthentication()},
		Bucket: "2093u4pkasjc3",
		Object: "flags.yaml",
	}
})
defer ffclient.Close()
```

## Configuration fields

To configure your Google Cloud Storage file location:

| Field        | Description                                                                                                                                                                    |
|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`Bucket`** | The name of your bucket.                                                                                                                                                       |
| **`Object`** | The name of your object in your bucket.                                                                                                                                        |
| **`Option`** | An instance of `option.ClientOption` that configures your access to Google Cloud. <br> Check [this documentation for more info](https://cloud.google.com/docs/authentication). |

