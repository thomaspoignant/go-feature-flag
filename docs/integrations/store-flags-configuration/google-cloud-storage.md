# Google Cloud Storage

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Retrieves the configuration from a Google Cloud Storage bucket. This retriever is useful when you are using Google Cloud and want to use GCS to store your configuration files.

<!-- -->

The [**Google Cloud Storage Retriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/gcstorageretriever/#Retriever) will use the [`google-cloud-storage package`](https://pkg.go.dev/cloud.google.com/go/storage) and [`google-api-options package`](https://pkg.go.dev/google.golang.org/api/option) to access your flag in Google Cloud Storage.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Google Cloud Storage<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: googleStorage

    bucket: 2093u4pkasjc3

    object: flags.yaml

# ...
```

| Field name | Mandatory | Type   | Default  | Description                                                                                                      |
| ---------- | --------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------- |
| `kind`     | **        | string | **none** | Value should be **`googleStorage`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `bucket`   | **        | string | **none** | This is the name of your Google Storage bucket *(ex: `my-featureflag-bucket`)*.                                  |
| `object`   | **        | string | **none** | Path to the file inside the bucket *(ex: `config/flag/my-flags.yaml`)*.                                          |

info

If you want to authenticate with Google Cloud, you can use the default environment variable `GOOGLE_APPLICATION_CREDENTIALS` to provide your credentials. We recommend using GCP native authentication methods.

Check [this documentation for more info](https://cloud.google.com/docs/authentication).

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Google Cloud Storage<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
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

| Field        | Mandatory | Description                                                                                                                                                                    |
| ------------ | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **`Bucket`** | **        | The name of your bucket.                                                                                                                                                       |
| **`Object`** | **        | The name of your object in your bucket.                                                                                                                                        |
| **`Option`** | **        | An instance of `option.ClientOption` that configures your access to Google Cloud.<br />Check [this documentation for more info](https://cloud.google.com/docs/authentication). |
