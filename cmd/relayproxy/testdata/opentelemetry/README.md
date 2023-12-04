# Test OpenTelemetry Tracing

**GO Feature Flag** is able to generate some trace if you use the OpenTelemetry.

If you want to test it locally, you can use the following command to start a local OpenTelemetry collector:

```bash
docker run --rm --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -e "COLLECTOR_OTLP_ENABLED=true" \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one
```

When your collector is up, you can configure GO Feature Flag relay-proxy by setting the OTLP endpoint in your configuration file.
```yaml
#...
openTelemetryOtlpEndpoint: http://localhost:4318
```

You can connect to **`jaeger`** at this address: http://localhost:16686/search.

After a 1st call to the API you will see a service call `go-feature-flag` and you can check the traces.
