# Webhook

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data by calling a HTTP Webhook.

<!-- -->

Everytime the `FlushInterval` or `MaxEventInMemory` is reached an HTTP call will be emitted to the <!-- -->Webhook<!-- --> exporter.

info

If for some reason the <!-- -->Webhook<!-- --> HTTP call fails, we will keep the data in memory and retry to add the next time we reach `FlushInterval` or `MaxEventInMemory`.

## Webhook format[​](#webhook-format "Direct link to Webhook format")

If you have configured a webhook, a `POST` request will be sent to the `EndpointURL` with a body in this format:

```
{

    "meta": {

        "hostname": "server01",

        // ...

    },

    "events": [

        {

            "kind": "feature",

            "contextKind": "anonymousUser",

            "userKey": "14613538188334553206",

            "creationDate": 1618909178,

            "key": "test-flag",

            "variation": "Default",

            "value": false,

            "default": false,

            "source": "SERVER"

        },

        // ...

    ]

}
```

## Signature[​](#signature "Direct link to Signature")

This header **`X-Hub-Signature-256`** is sent if the webhook is configured with a **`secret`**. This is the **HMAC hex digest** of the request body, and is generated using the **SHA-256** hash function and the **secret as the HMAC key**.

caution

The recommendation is to always use the `Secret` and on your API/webhook always verify the signature key to be sure that you don't get into a man-in-the-middle attack.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Webhook<!-- --> exporter, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

exporters:

  - kind: webhook

    endpointUrl: https://your-webhook-url.com/

# ...
```

| Field name         | Mandatory | Type                  | Default  | Description                                                                                                                                                    |
| ------------------ | --------- | --------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`             | **        | string                | **none** | **Value should be `webhook`**.<br />*This field is mandatory and describes which retriever you are using.*                                                     |
| `endpointUrl`      | **        | string                | **none** | EndpointURL of your webhook.                                                                                                                                   |
| `flushInterval`    | **        | int                   | `60000`  | The interval in millisecond between 2 calls to the webhook *(if the `maxEventInMemory` is reached before the flushInterval we will call the exporter before)*. |
| `maxEventInMemory` | **        | int                   | `100000` | If we hit that limit we will call the exporter.                                                                                                                |
| `secret`           | **        | string                | **none** | Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br />See [signature section](#signature) for more details.                    |
| `meta`             | **        | map\[string]string    | **none** | Add all the information you want to see in your request.                                                                                                       |
| `headers`          | **        | map\[string]\[]string | **none** | Add all the headers you want to add while calling the endpoint                                                                                                 |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Webhook<!-- --> exporter, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
config := ffclient.Config{

    // ...

   DataExporter: ffclient.DataExporter{

        // ...

        Exporter: &webhookexporter.Exporter{

            EndpointURL: " https://webhook.url/",

            Secret:      "secret-for-signing",

            Meta:        map[string]string{

                "extraInfo": "info",

            },

            Headers: map[string][]string{

                "Authorization": {"Bearer auth_token"},

            },

        },

    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field         | Mandatory | Description                                                                                                                                 |
| ------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `EndpointURL` | **        | EndpointURL of your webhook                                                                                                                 |
| `Secret`      | **        | Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br />See [signature section](#signature) for more details. |
| `Meta`        | **        | Add all the information you want to see in your request.                                                                                    |
| `Headers`     | **        | List of Headers to send to the endpoint                                                                                                     |
