---
sidebar_position: 7
---

# MongoDB
The `mongodbRetriever` will use the mongoDB database to get your flags.

## Example
```go linenums="1"
awsConfig, _ := config.LoadDefaultConfig(context.Background())
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
	  Retriever: &mongodbretriever.Retriever{
        Collection: "featureFlags",
        Database: "appConfig",
        URI: "mongodb://root:example@127.0.0.1:27017/",
    },
})
defer ffclient.Close()
```

## Expected format
If you use MongoDB to store your flags, you need a specific format to store your flags.

We expect the flag to be stored in JSON format as defined in the [flag format](../../configure_flag/flag_format#format-details),
but you should also add a new field called `flag` containing the name of the flag.

The retriever will read all the flags from the collection.

### Example:
```json
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

## Configuration fields
To configure your mongodb retriever:

| Field            | Description                                                 |
|------------------|-------------------------------------------------------------|
| **`Collection`** | Name of the collection where your flags are stored          |
| **`Database`**   | Name of the mongo database where the collection is located. |
| **`URI`**        | Connection URI of your mongoDB instance.                    |
