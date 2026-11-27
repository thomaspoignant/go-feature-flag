# Azure Blob Storage

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Retrieves the configuration from an Azure Blob Storage. This retriever is useful when you are using Azure and want to use Azure Blob Storage to store your configuration files.

<!-- -->

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Azure Blob Storage<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: azureBlobStorage

    container: my-featureflag-container

    accountName: my-featureflag-account

    object: flag/flags.goff.yaml

# ...
```

| Field name    | Mandatory | Type   | Default  | Description                                                                                                         |
| ------------- | --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------------------- |
| `kind`        | **        | string | **none** | Value should be **`azureBlobStorage`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `container`   | **        | string | **none** | This is the name of your Azure Blob Storage container *(ex: `my-featureflag-container`)*.                           |
| `accountName` | **        | string | **none** | This is the name of your Azure Blob Storage account.                                                                |
| `object`      | **        | string | **none** | Location of your configuration file.                                                                                |
| `accountKey`  | **        | string | **none** | This is the secret key of your Azure Blob Storage account.                                                          |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Azure Blob Storage<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
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

| Field             | Mandatory | Description                                    |
| ----------------- | --------- | ---------------------------------------------- |
| **`Container`**   | **        | Name of the Azure Blob Storage container       |
| **`Object`**      | **        | Name of the feature flag file in the container |
| **`AccountName`** | **        | Azure Storage Account Name                     |
| **`AccountKey`**  | **        | Storage Account Key                            |
| **`ServiceURL`**  | **        | Custom service URL                             |

## Error Handling[​](#error-handling "Direct link to Error Handling")

The `Retrieve` method returns an error if:

* `AccountName` is empty
* `Container` or `Object` is not specified
* There's an issue initializing the Azure client.
* There's a problem downloading or reading the file from Azure Blob Storage.
