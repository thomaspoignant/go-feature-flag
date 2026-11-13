# Webhook

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Send notifications to a Webhook in a specific format.

<!-- -->

## About <!-- -->Webhook<!-- --> Notifier[​](#about-notifiername-notifier "Direct link to about-notifiername-notifier")

The **Webhook notifier** will perform an HTTP POST request to the specified endpoint everytime a change in the flags is detected.

The format of the call is specified in the [format section](#format) and you can [sign the body](#signature) to trust the data you are receiving.

## Format[​](#format "Direct link to Format")

If you have configured a webhook, a `POST` request will be sent to the `EndpointURL` with a body in this format:

```
{

  "meta": {

    "hostname": "server01"

    // ...

  },

  "flags": {

    "deleted": {},

    // map of your deleted flags

    "added": {},

    // map of your added flags

    "updated": {

      "flag-name": {

        // an object that contains old and new value

        "old_value": {},

        "new_value": {}

      }

    }

  }

}
```

### Example[​](#example "Direct link to Example")

```
{

  "meta": {

    "hostname": "server01"

  },

  "flags": {

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

## Signature[​](#signature "Direct link to Signature")

This header **`X-Hub-Signature-256`** is sent if the webhook is configured with a secret. This is the HMAC hex digest of the request body, and is generated using the SHA-256 hash function and the secret as the HMAC key.

Notifications can be delivered concurrently. Verify each request's signature against that request's body using your configured secret.

caution

The recommendation is to always use the `Secret` and on your API/webook always verify the signature key to be sure that you don't get into a man-in-the-middle attack.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Webhook<!-- --> notifier, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

notifier:

  - kind: webhook

    endpointUrl: "https://example.com/hook"

# ...
```

| Field name    | Mandatory | Type                  | Default  | Description                                                                                                                                                                                                                 |
| ------------- | --------- | --------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`        | **        | string                | **none** | **Value should be `webhook`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                                  |
| `endpointUrl` | **        | string                | **none** | The complete URL of your API (we will send a POST request to this URL, see [format](https://thomaspoignant.github.io/go-feature-flag/latest/notifier/webhook/#format)                                                       |
| `secret`      | **        | string                | **none** | Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br />See [signature section](https://thomaspoignant.github.io/go-feature-flag/latest/data_collection/webhook/#signature) for more details. |
| `meta`        | **        | map\[string]string    | **none** | Add all the information you want to see in your request.                                                                                                                                                                    |
| `headers`     | **        | map\[string]\[]string | **none** | Add all the headers you want to add while calling the endpoint                                                                                                                                                              |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Webhook<!-- --> notifier, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

    // ...

    Notifiers: []notifier.Notifier{

        &webhooknotifier.Notifier{

            EndpointURL: "https://example.com/hook",

            Secret:     "Secret",

            Meta: map[string]string{

                "app.name": "my app",

            },

            Headers: map[string][]string{

                "Authorization": {"Bearer auth_token"},

            },

        },

    // ...

    },

})

defer ffclient.Close()
```

| Field         | Mandatory | Description                                                                                                                                                                                                                              |
| ------------- | --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `EndpointURL` | **        | The complete URL of your API *(we will send a POST request to this URL, [see format](#format))*                                                                                                                                          |
| `Secret`      | **        | A secret key you can share with your webhook. We will use this key to sign the request *(see [signature section](#signature) for more details)*.                                                                                         |
| `Meta`        | **        | A list of key value that will be added in your request, this is super useful if you want to add information on the current running instance of your app.<br /><br />**By default the hostname is always added in the meta information.** |
| `Headers`     | **        | The list of Headers to send to the endpoint.                                                                                                                                                                                             |
