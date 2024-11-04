# Kafka exporter example

This example contains everything you need to export the usage of your flag to **`AWS Kinesis`**.

## How to setup the example
_All commands should be run in the root level of the repository._

1. Start a kafka server by running:

```shell
docker-compose -f examples/data_export_kinesis/docker-compose.yml up
```

2. Create python virtualenv & Install awslocal
```shell
mkvirtualenv go-features-flag-kinesis
# or activate existent
workon go-features-flag-kinesis
pip install awscli awscli-local
```

2. Create a topic in Kinesis:

```shell
awslocal kinesis create-stream --stream-name test-stream --region us-east-1
```

3. Update dependencies:

```shell
make vendor
```

4. Run the example application:

```shell
go run ./examples/data_export_kinesis/main.go
```
_If you check the logs, you should see the events being sent 1 by 1 to kinesis._

5. Read the items in the topic:

```shell
SHARD_ITERATOR=$(awslocal kinesis get-shard-iterator --shard-id shardId-000000000000 --shard-iterator-type TRIM_HORIZON --stream-name test-stream --query 'ShardIterator' --region us-east-1)

awslocal kinesis get-records --shard-iterator $SHARD_ITERATOR --region us-east-1
```
