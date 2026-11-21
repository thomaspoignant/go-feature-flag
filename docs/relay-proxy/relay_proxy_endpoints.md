# 🌐 API endpoints

## Overview[​](#overview "Direct link to Overview")

The relay-proxy is providing a list of REST endpoints to interact with GO Feature Flag.

Since we are evolving the relay proxy, the list of endpoints can change, and we recommend checking the [**OpenAPI documentation**](https://github.com/thomaspoignant/go-feature-flag/blob/main/cmd/relayproxy/docs/swagger.json) to have the most updated list of endpoints.

### Swagger[​](#swagger "Direct link to Swagger")

The relay proxy can expose a [swagger UI](https://swagger.io/tools/swagger-ui/) to test your REST endpoints.

This swagger UI is available at the `/swagger/` endpoint and will show all the available endpoints and the parameters needed.

info

By default, this endpoint is not exposed, you need to have this configuration in your **relay proxy** configuration file:

goff-proxy.yaml

```
# ...

enableSwagger: true

host: my-proxy-domain.com # the DNS to access the proxy
```

When enabled, you can go to the `/swagger/` endpoint with your browser, and you will have access to the Swagger UI for the relay proxy.

## [OpenAPI documentation](https://gofeatureflag.org/API_relayproxy)[​](#openapi-documentation "Direct link to openapi-documentation")

If you don't want to install the relay proxy to check the endpoints, you can go to this [**OpenAPI documentation**](https://gofeatureflag.org/API_relayproxy) directly.

## Flag change streams[​](#flag-change-streams "Direct link to Flag change streams")

The relay proxy can push a message to your application each time the flag configuration changes.

| Transport | Endpoint                     | Notes                                                  |
| --------- | ---------------------------- | ------------------------------------------------------ |
| WebSocket | `/stream/v1/ws/flag/change`  | Receives the diff of the changed flags.                |
| WebSocket | `/ws/v1/flag/change`         | **Deprecated** alias of `/stream/v1/ws/flag/change`.   |
| SSE       | `/stream/v1/sse/flag/change` | Receives a `refetchEvaluation` event (OFREP ADR-0008). |

The API key is passed with the `apiKey` query parameter, for example `ws://localhost:1031/stream/v1/ws/flag/change?apiKey=<your-api-key>`.

### Streams with flag sets[​](#streams-with-flag-sets "Direct link to Streams with flag sets")

When the relay proxy runs with [flag sets](https://gofeatureflag.org/docs/concepts/flagset.md), each stream is scoped to a single flag set:

* The connection only receives the changes of the flag set associated to its API key.
* The API key **must be associated to a flag set**. A missing or unknown key is rejected with `401`, and a valid key that is not associated to any flag set (for example a key from `authorizedKeys`) is rejected with `400`, before the WebSocket upgrade.
* To watch several flag sets, open **one connection per flag set**, each with an API key of that flag set.

warning

Older versions of the relay proxy sent the changes of every flag set to every WebSocket client. If you relied on this behavior, open one connection per flag set. If you still use the deprecated `/ws/v1/flag/change` endpoint, it follows the same rules; migrate to `/stream/v1/ws/flag/change` when you update your clients.

Without flag sets, nothing changes: every connection receives all flag changes.
