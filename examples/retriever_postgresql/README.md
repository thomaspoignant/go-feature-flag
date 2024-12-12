# PostgreSQL example

This example contains everything you need to use **`PostgreSQL`** as the source for your flags.

As you can see the `main.go` file contains a basic HTTP server that expose an API that use your flags.

## How to setup the example
_All commands should be run in the root level of the repository._

1. Load all dependencies

```shell
make vendor
```

2. Run the PostgreSQL container provided in the `docker-compose.yml` file.

```shell
docker compose -f ./example/retriever_postgresql/docker-compose.yml up
```

The container will run an initialization script that populates the database with example flags

3. Run the example app to visualize the flags being evaluated

```shell
go run ./examples/retriever_postgresql/main.go
```

4. Play with the values in the configured MongoDB documents to see different outputs
