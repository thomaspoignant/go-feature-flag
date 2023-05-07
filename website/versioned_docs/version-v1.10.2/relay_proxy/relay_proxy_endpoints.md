---
sidebar_position: 60
description: Description of the available endpoints in the relay proxy.
---

# Relay proxy endpoints

The most updated documentation about the relay proxy endpoints is the Swagger docs _(see [Swagger section](#swagger) to see how to access to the documentation)_.

### Swagger
Swagger endpoint is serving a [swagger UI](https://swagger.io/tools/swagger-ui/) to test your REST endpoints.
By default, this endpoint is not exposed, you need to have this configuration in your **relay proxy** configuration file:

```yaml
# ...

enableSwagger: true
host: my-proxy-domain.com # the DNS to access the proxy
```

When enabled, you can go to the `/swagger/` endpoint with your browser, and you will have access to the Swagger UI for the relay proxy. 

## [OpenAPI documentation](/API_relayproxy)

If you don't want to install the relay proxy to check the endpoints, you can go to this [**OpenAPI documentation**](/API_relayproxy) directly.
