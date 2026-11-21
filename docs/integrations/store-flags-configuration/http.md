# HTTP(S)

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Fetches the configuration from a remote URL over HTTP(S). This retriever is useful when you want to fetch the configuration from a remote server endpoint.

<!-- -->

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->HTTP(S)<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: http

    url: https://my-feature-flag-location.com/flags.goff.yaml

# ...
```

| Field name       | Mandatory | Type   | Default  | Description                                                                                             |
| ---------------- | --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------- |
| `kind`           | **        | String | **none** | **Value should be `http`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `url`            | **        | String | **none** | Location to retrieve the file.                                                                          |
| `method`         | **        | String | `GET`    | The HTTP Method you are using to call the HTTP endpoint.                                                |
| `body`           | **        | String | **none** | The HTTP Body you are using to call the HTTP endpoint.                                                  |
| `headers`        | **        | Object | **none** | The HTTP headers used when calling the HTTP endpoint (useful for authorization).                        |
| `clientCertPath` | **        | String | **none** | Path to the client certificate file used for mTLS. Must be configured with `clientKeyPath`.             |
| `clientKeyPath`  | **        | String | **none** | Path to the client certificate key file used for mTLS. Must be configured with `clientCertPath`.        |
| `caCertPath`     | **        | String | **none** | Path to the CA certificate file used to verify the server certificate.                                  |
| `timeout`        | **        | String | `10000`  | Timeout in millisecond when calling the HTTP endpoint.                                                  |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->HTTP(S)<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &httpretriever.Retriever{

        URL:    "http://example.com/flag-config.goff.yaml",

        Timeout: 2 * time.Second,

    },

})

defer ffclient.Close()
```

| Field              | Mandatory | Description                                                                                      |
| ------------------ | --------- | ------------------------------------------------------------------------------------------------ |
| **URL**            | **        | Location to retrieve the file<br />*(ex: <http://mydomain.io/flag.yaml>)*.                       |
| **Method**         | **        | the HTTP method you want to use<br />*(default is `GET`)*.                                       |
| **Body**           | **        | If you need a body to get the flags.                                                             |
| **Header**         | **        | Header you should pass while calling the endpoint *(useful for authorization)*.                  |
| **ClientCertPath** | **        | Path to the client certificate file used for mTLS. Must be configured with `ClientKeyPath`.      |
| **ClientKeyPath**  | **        | Path to the client certificate key file used for mTLS. Must be configured with `ClientCertPath`. |
| **CACertPath**     | **        | Path to the CA certificate file used to verify the server certificate.                           |
| **Timeout**        | **        | Timeout for the HTTP call<br />(default is 10 seconds).                                          |
