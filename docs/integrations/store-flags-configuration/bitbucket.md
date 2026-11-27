# Bitbucket

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Fetch the configuration from files stored in a GIT repository.

<!-- -->

tip

Bitbucket has rate limits, be sure to correctly set your `PollingInterval` to avoid reaching the limit.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Bitbucket<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: bitbucket

    repositorySlug: thomaspoignant/go-feature-flag

    path: config/flag/my-flags.yaml

# ...
```

| Field name       | Mandatory | Type   | Default              | Description                                                                                                                                                                       |
| ---------------- | --------- | ------ | -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`           | **        | string | **none**             | Value should be **`bitbucket`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                      |
| `repositorySlug` | **        | string | **none**             | The repository slug of the Bitbucket repository where your file is located *(ex: `thomaspoignant/go-feature-flag`)*.                                                              |
| `path`           | **        | string | **none**             | Path to the file inside the repository *(ex: `config/flag/my-flags.yaml`)*.                                                                                                       |
| `baseUrl`        | **        | string | `https://gitlab.com` | The base URL of your Bitbucket instance<br />By default we are using the public API `https://api.bitbucket.org`.                                                                  |
| `branch`         | **        | string | `main`               | The branch we should check in the repository.                                                                                                                                     |
| `token`          | **        | string | **none**             | Bitbucket token used to access a private repository ([*Create a Repository Access Token*](https://support.atlassian.com/bitbucket-cloud/docs/create-a-repository-access-token/)). |
| `timeout`        | **        | string | `10000`              | Timeout in millisecond used when calling GitLab.                                                                                                                                  |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Bitbucket<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &bitbucketretriever.Retriever{

        RepositorySlug: "thomaspoignant/go-feature-flag",

        Branch: "main",

        FilePath: "testdata/flag-config.goff.yaml",

        BitBucketToken: "XXXX",

        Timeout: 2 * time.Second,

    },

})

defer ffclient.Close()
```

| Field                | Mandatory | Description                                                                          |
| -------------------- | --------- | ------------------------------------------------------------------------------------ |
| **`RepositorySlug`** | **        | Your Gitlab slug `org/repo-name`.                                                    |
| **`FilePath`**       | **        | The path of your file.                                                               |
| **`BaseURL`**        | **        | The domain name of your Bitbucket instance<br />Default: `https://api.bitbucket.org` |
| **`Branch`**         | **        | The branch where your file is.<br />Default: `main`                                  |
| **`BitBucketToken`** | **        | Bitbucket token is used to access a private repository                               |
| **`Timeout`**        | **        | Timeout for the HTTP call<br />Default: 10 seconds                                   |
