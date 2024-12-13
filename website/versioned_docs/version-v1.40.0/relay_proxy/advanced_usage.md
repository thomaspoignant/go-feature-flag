---
sidebar_position: 90
title: Advanced usage
description: All the advanced usage of the relay proxy.
---

## Manually trigger retrievers and refresh the internal flag cache
By default, the retrievers are called regularly to refresh the configuration based on the polling interval.

But there are use cases where you want to refresh the configuration explicitly _(for example, during the CI process
after you have changed your configuration file)_.

To do so you can call the `/v1/admin/retriver/refresh` endpoint with a POST request.  
It will force the retrievers to be called and update the internal cache.

```shell
curl -X 'POST' \
  'http://<your_domain>:1031/admin/v1/retriever/refresh' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer <your_admin_api_key>' \
  -d ''
```

:::note
This endpoint must be called with an admin token.
Authorized keys should be configured in the relay-proxy configuration file under the key `authorizedKeys.admin`.
:::

