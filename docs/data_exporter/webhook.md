# Webhook Exporter

The **Webhook exporter** will collect the data and send them via an HTTP POST request to the specified endpoint.
Everytime the `FlushInterval` or `MaxEventInMemory` is reached a new call is performed.  

!!! Info
    If for some reason the call failed, we will keep the data in memory and retry to add the next time we reach `FlushInterval` or `MaxEventInMemory`.

## Configuration example
```go linenums="1"
ffclient.Config{ 
    // ...
   DataExporter: ffclient.DataExporter{
        // ...
        Exporter: &ffexporter.Webhook{
            EndpointURL: " https://webhook.url/",
            Secret:      "secret-for-signing",
            Meta:        map[string]string{
                "extraInfo": "info",
            },
        },
    },
    // ...
}
```
## Configuration fields
| Field  | Description  |
|---|---|
|`EndpointURL `   | EndpointURL of your webhook |
|`Secret `   |  *(optional)*<br>Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br>See [signature section](#signature) for more details.  |
|`Meta`   |   *(optional)*<br>Add all the information you want to see in your request. |


## Webhook format
If you have configured a webhook, a `POST` request will be sent to the `EndpointURL` with a body in this format:

```json linenums="1"
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
            "default": false
        },
        // ...
    ]
}
```

## Signature
This header **`X-Hub-Signature-256`** is sent if the webhook is configured with a **`secret`**.  
This is the **HMAC hex digest** of the request body, and is generated using the **SHA-256** hash function and the **secret as the HMAC key**.

!!! Danger
    The recommendation is to always use the `Secret` and on your API/webook always verify the signature key to be sure that you don't have a man in the middle attack.
