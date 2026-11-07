# MongoDB

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Load the configuration from a MongoDB collection. This retriever is useful when you are using MongoDB and want to use a collection to store your configuration files.

<!-- -->

## MongoDB Collection Format[​](#mongodb-collection-format "Direct link to MongoDB Collection Format")

If you use MongoDB to store your flags, you need a specific format to store your flags.

We expect the flag to be stored in JSON format as defined in the [create flags section](https://gofeatureflag.org/docs/configure_flag/create-flags.md), but you should also add a new field called `flag` containing the name of the flag.

The retriever will read all the flags from the collection at once.

#### Example[​](#example "Direct link to Example")

mongoDB flag format

```
{

    "flag": "new-admin-access",

    "variations": {

        "default_var": false,

        "false_var": false,

        "true_var": true

    },

    "defaultRule": {

        "percentage": {

            "false_var": 70,

            "true_var": 30

        }

    }

}
```

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->MongoDB<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: mongodb

    uri: mongodb://root:example@127.0.0.1:27017/

    database: appConfig

    collection: featureFlags

# ...
```

| Field name   | Mandatory | Type   | Default  | Description                                                                                                |
| ------------ | --------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------- |
| `kind`       | **        | string | **none** | Value should be **`mongodb`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `uri`        | **        | string | **none** | This is the MongoDB URI used in order to connect to the MongoDB instance.                                  |
| `database`   | **        | string | **none** | Name of the **database** where flags are stored.                                                           |
| `collection` | **        | string | **none** | Name of the **collection** where flags are stored.                                                         |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->MongoDB<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

  PollingInterval: 3 * time.Second,

  Retriever: &mongodbretriever.Retriever{

    Collection: "featureFlags",

    Database:   "appConfig",

    URI:        "mongodb://root:example@127.0.0.1:27017/",

  },

})

defer ffclient.Close()
```

| Field            | Mandatory | Description                                                 |
| ---------------- | --------- | ----------------------------------------------------------- |
| **`Collection`** | **        | Name of the collection where your flags are stored          |
| **`Database`**   | **        | Name of the mongo database where the collection is located. |
| **`URI`**        | **        | Connection URI of your mongoDB instance.                    |
