---
sidebar_position: 7
---

# Redis
The `redisRetriever` will use the redis database to get your flags.

## Example
```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
	  Retriever: &redisRetriver.Retriever{
        Prefix: "goff:",
		    Options: &redis.Options{
            Addr: "127.0.0.1:6379",
        },
    },
})
defer ffclient.Close()
```

## Expected format
If you use Redis to store your flags, you need a specific format to store your flags.

We expect the flag to be stored as a `string:string` format where the key if the flag key (with or without a prefix) 
and the value is a string representing the flag in JSON.

The retriever will `Scan` redis filtering with the `Prefix` and will parse the value as a JSON object.
`
## Configuration fields
To configure your redis retriever:

| Field         | Description                                                                           |
|---------------|---------------------------------------------------------------------------------------|
| **`Options`** | A `redis.Options` object containing the connection information to the redis instance. |
| **`Prefix`**  | (optional) Key prefix to filter on the key names.                                     |
