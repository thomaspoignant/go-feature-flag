# Redis

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Load the configuration from Redis using a specific prefix.

<!-- -->

## Redis Format[​](#redis-format "Direct link to Redis Format")

If you use Redis to store your flags, you need a specific format to store your flags.

We expect the flag to be stored as a `string:string` format where the key is the flag key (with or without a prefix) and the value is a string representing the flag in JSON.

The retriever will `Scan` redis filtering with the `Prefix` and will parse the value as a `JSON object`.

The retriever will read all the flags from Redis at once.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Redis<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: redis

    options:

      addr: "127.0.0.1:6379"

    prefix: "goff:"

# ...
```

| Field name | Mandatory | Type   | Default  | Description                                                                                                                                                                                                                            |
| ---------- | --------- | ------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`     | **        | string | **none** | **Value should be `redis`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                                               |
| `options`  | **        | object | **none** | Options used to connect to your redis instance.<br />All the options from the `go-redis` SDK are available *([check `redis.Options`](https://github.com/redis/go-redis/blob/683f4fa6a6b0615344353a10478548969b09f89c/options.go#L31))* |
| `prefix`   | **        | string | **none** | Prefix used before your flag name in the Redis DB.                                                                                                                                                                                     |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Redis<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

	  Retriever: &redisretriever.Retriever{

        Prefix: "goff:",

		    Options: &redis.Options{

            Addr: "127.0.0.1:6379",

        },

    },

})

defer ffclient.Close()
```

| Field         | Mandatory | Description                                                                                                                                      |
| ------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **`Options`** | **        | A [`redis.Options`](https://pkg.go.dev/github.com/redis/go-redis/v9#Options) object containing the connection information to the redis instance. |
| **`Prefix`**  | **        | Key prefix to filter on the key names.                                                                                                           |
