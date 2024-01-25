# MongoDE example

This example contains everything you need to use **`MongoDB`** as the source for your flags.

As you can see the `main.go` file contains a basic HTTP server that expose an API that use your flags.

## How to setup the example
_All commands should be run in the root level of the repository._

1. Load all dependencies

```shell
make vendor
```

2. Run the MongoDB container provided in the `docker-compose.yml` file:

```shell
docker compose -f ./example/retriever_mongodb/docker-compose.yml up
```

3. Connect to the MongoDB container using MongoDB Compass or any other tool using the following connection string: `mongodb://root:example@127.0.0.1:27017/`

4. Create the `appConfig` database and the `featureFlags` collection within it. After this, insert a JSON flag definition below, where the `flag` key has the feature flag name as value.

```json
[
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
    },
    {
        "flag": "flag-only-for-admin",
        "variations": {
            "default_var": false,
            "false_var": false,
            "true_var": true
        },
        "targeting": [
            {
                "query": "admin eq true",
                "percentage": {
                    "false_var": 0,
                    "true_var": 100
                }
            }
        ],
        "defaultRule": {
            "variation": "Default"
        }
    }
]
```

5. Build the relay proxy

```shell
make build-relayproxy
```

6. Execute the relay proxy with the example configuration

```shell
./out/bin/relayproxy --config ./examples/retriever_mongodb/mongo-retriever-config.yam
```

7. Run the example app to visualize the flags being evaluated

```shell
go run ./examples/retriever_mongodb/main.go
```

8. Play with the values in the configured MongoDB documents to see different outputs
