# GitHub

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Fetch the configuration from files stored in a GitHub repository. This retriever will perform an HTTP Request with your GitHub configuration on the GitHub API to get your flags.

<!-- -->

info

GitHub has rate limits, be sure to correctly set your `PollingInterval` to avoid reaching the limit.

**If the rate limit is reached, the retriever will log an error and will stop polling until GitHub allows it again.**

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->GitHub<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: github

    repositorySlug: thomaspoignant/go-feature-flag

    path: config/flag/my-flags.yaml

# ...
```

| Field name       | Mandatory | Type   | Default  | Description                                                                                                                                                                                                                        |
| ---------------- | --------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`           | **        | string | **none** | **Value should be `github`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                                          |
| `repositorySlug` | **        | string | **none** | The repository slug of the GitHub repository where your file is located *(ex: `thomaspoignant/go-feature-flag`)*.                                                                                                                  |
| `path`           | **        | string | **none** | Path to the file inside the repository *(ex: `config/flag/my-flags.yaml`)*.                                                                                                                                                        |
| `branch`         | **        | string | `main`   | The branch we should check in the repository.                                                                                                                                                                                      |
| `token`          | **        | string | **none** | Github token used to access a private repository, you need the repo permission ([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)). |
| `baseUrl`        | **        | string | `none`   | The base URL for the GitHub API. Use this for GitHub Enterprise or self-hosted instances *(ex: `https://github.acme.com/api/v3`)*. If not specified, defaults to `https://api.github.com`.                                         |
| `timeout`        | **        | string | `10000`  | Timeout in millisecond used when calling GitHub.                                                                                                                                                                                   |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->GitHub<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &githubretriever.Retriever{

        RepositorySlug: "thomaspoignant/go-feature-flag",

        Branch: "main",

        FilePath: "testdata/flag-config.goff.yaml",

        GithubToken: "XXXX",

        Timeout: 2 * time.Second,

    },

})

defer ffclient.Close()
```

| Field                | Mandatory | Description                                                                                                                                                                                                                               |
| -------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`RepositorySlug`** | **        | Your GitHub slug `org/repo-name`.                                                                                                                                                                                                         |
| **`FilePath`**       | **        | The path of your file.                                                                                                                                                                                                                    |
| **`Branch`**         | **        | The branch where your file is.<br />Default: `main`                                                                                                                                                                                       |
| **`GithubToken`**    | **        | Github token is used to access a private repository, you need the `repo` permission *([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token))*. |
| **`BaseURL`**        | **        | The base URL for the GitHub API. Use this for GitHub Enterprise or self-hosted instances *(ex: `https://github.acme.com/api/v3`)*. If not specified, defaults to `https://api.github.com`.                                                |
| **`Timeout`**        | **        | Timeout for the HTTP call<br />Default: 10 seconds                                                                                                                                                                                        |

## GitHub Enterprise / Self-hosted instances[​](#github-enterprise--self-hosted-instances "Direct link to GitHub Enterprise / Self-hosted instances")

If you are using a GitHub Enterprise or self-hosted instance, you can configure the `baseUrl` field to point to your instance's API endpoint.

### Relay Proxy Configuration[​](#relay-proxy-configuration "Direct link to Relay Proxy Configuration")

goff-proxy.yaml

```
# ...

retrievers:

  - kind: github

    repositorySlug: myorg/myrepo

    path: config/flag/my-flags.yaml

    baseUrl: "https://github.acme.com/api/v3"

# ...
```

### GO Module Configuration[​](#go-module-configuration "Direct link to GO Module Configuration")

example.go

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &githubretriever.Retriever{

        RepositorySlug: "myorg/myrepo",

        Branch: "main",

        FilePath: "config/flag/my-flags.yaml",

        BaseURL: "https://github.acme.com/api/v3",

        GithubToken: "XXXX",

    },

})

defer ffclient.Close()
```
