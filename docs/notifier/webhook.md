# Webhook Notifier
The **Webhook notifier** will perform an HTTP POST request to the specified endpoint everytime that a change in the flags is detected.

The format of the call is specified in the [format section](#format) and, you can [sign the body](#signature) to trust the data you are receiving. 

## Configure the webhook notifier

```go linenums="1" 
ffclient.Config{ 
    // ...
    Notifiers: []ffclient.NotifierConfig{
        &ffclient.WebhookConfig{
            EndpointURL: " https://example.com/hook",
            Secret:     "Secret",
            Meta: map[string]string{
                "app.name": "my app",
            },
        },
        // ...
    },
}
```

## Configuration fields
| Field  | Description   |
|---|---|
|`EndpointURL`   | The complete URL of your API *(we will send a POST request to this URL, [see format](#format))*  |
|`Secret`   |  *(optional)*<br>A secret key you can share with your webhook. We will use this key to sign the request *(see [signature section](#signature) for more details)*. |
|`Meta`   |  *(optional)*<br>A list of key value that will be add in your request, this is super useful if you want to add information on the current running instance of your app.<br/><br/>**By default the hostname is always added in the meta information.**|

## Format
If you have configured a webhook, a `POST` request will be sent to the `EndpointURL` with a body in this format:

```json linenums="1"
{
    "meta": {
        "hostname": "server01",
        // ...
    },
    "flags": {
        "deleted": {}, // map of your deleted flags
        "added": {}, // map of your added flags
        "updated": {
            "flag-name": { // an object that contains old and new value
                "old_value": {},
                "new_value": {}
            }
        }
    }
}
```

### Example

```json linenums="1"
{
   "meta":{
       "hostname": "server01"
   },
   "flags":{
       "deleted": {
           "test-flag": {
               "rule": "key eq \"random-key\"",
               "percentage": 100,
               "true": true,
               "false": false,
               "default": false
           }
       },
       "added": {
           "test-flag3": {
               "percentage": 5,
               "true": "test",
               "false": "false",
               "default": "default"
           }
       },
       "updated": {
           "test-flag2": {
               "old_value": {
                   "rule": "key eq \"not-a-key\"",
                   "percentage": 100,
                   "true": true,
                   "false": false,
                   "default": false
               },
               "new_value": {
                   "disable": true,
                   "rule": "key eq \"not-a-key\"",
                   "percentage": 100,
                   "true": true,
                   "false": false,
                   "default": false
               }
           }
       }
   }
}
```

## Signature
This header **`X-Hub-Signature-256`** is sent if the webhook is configured with a secret. This is the HMAC hex digest of the request body, and is generated using the SHA-256 hash function and the secret as the HMAC key.

!!! Danger
    The recommendation is to always use the `Secret` and on your API/webook always verify the signature key to be sure that you don't have a man in the middle attack.
