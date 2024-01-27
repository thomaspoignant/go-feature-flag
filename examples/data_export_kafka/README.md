# Kafka exporter example

This example contains everything you need to export the usage of your flag to **`Kafka`**.

## How to setup the example
_All commands should be run in the root level of the repository._

1. Start a kafka server by running:

```shell
docker-compose -f examples/data_export_kafka/docker-compose.yml up
```

2. Create a topic in Kafka:

```shell
docker exec $(docker ps | grep cp-kafka |  awk '{print $1}') kafka-topics --create --topic go-feature-flag-events --bootstrap-server localhost:9092
```
3. Update dependencies:

```shell
make vendor
```

4. Run the example application:

```shell
go run ./examples/data_export_kafka/main.go
```
_If you check the logs, you should see the events being sent 1 by 1 to kafka._

5. Read the items in the topic:

```shell
docker exec $(docker ps | grep cp-kafka |  awk '{print $1}')  kafka-console-consumer --bootstrap-server localhost:9092 --topic go-feature-flag-events --from-beginning
```
