---
sidebar_position: 60
description: Description of the available endpoints in the relay proxy.
---

# 🌐 API endpoints

## Overview
The relay-proxy is providing a list of REST endpoints to interact with GO Feature Flag.

Since we are evolving the relay proxy, the list of endpoints can change, and we recommend checking the [**OpenAPI documentation**](https://github.com/thomaspoignant/go-feature-flag/blob/main/cmd/relayproxy/docs/swagger.json) to have the most updated list of endpoints.

### Swagger
The relay proxy can expose a [swagger UI](https://swagger.io/tools/swagger-ui/) to test your REST endpoints.

This swagger UI is available at the `/swagger/` endpoint and will show all the available endpoints and the parameters needed.

:::info
By default, this endpoint is not exposed, you need to have this configuration in your **relay proxy** configuration file:

```yaml title="goff-proxy.yaml"
# ...
enableSwagger: true
host: my-proxy-domain.com # the DNS to access the proxy
```

When enabled, you can go to the `/swagger/` endpoint with your browser, and you will have access to the Swagger UI for the relay proxy. 
:::

## [OpenAPI documentation](/API_relayproxy)

If you don't want to install the relay proxy to check the endpoints, you can go to this [**OpenAPI documentation**](/API_relayproxy) directly.
