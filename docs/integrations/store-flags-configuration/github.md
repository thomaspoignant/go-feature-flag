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

| Field name                | Mandatory | Type   | Default  | Description                                                                                                                                                                                                                                                                                   |
| ------------------------- | --------- | ------ | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`                    | **        | string | **none** | **Value should be `github`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                                                                                                     |
| `repositorySlug`          | **        | string | **none** | The repository slug of the GitHub repository where your file is located *(ex: `thomaspoignant/go-feature-flag`)*.                                                                                                                                                                             |
| `path`                    | **        | string | **none** | Path to the file inside the repository *(ex: `config/flag/my-flags.yaml`)*.                                                                                                                                                                                                                   |
| `branch`                  | **        | string | `main`   | The branch we should check in the repository.                                                                                                                                                                                                                                                 |
| `token`                   | **        | string | **none** | GitHub token (PAT) used to access a private repository, you need the repo permission ([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)). Mutually exclusive with the GitHub App fields below. |
| `baseUrl`                 | **        | string | `none`   | The base URL for the GitHub API. Use this for GitHub Enterprise or self-hosted instances *(ex: `https://github.acme.com/api/v3`)*. If not specified, defaults to `https://api.github.com`.                                                                                                    |
| `timeout`                 | **        | string | `10000`  | Timeout in millisecond used when calling GitHub.                                                                                                                                                                                                                                              |
| `githubAppId`             | **        | int    | **none** | GitHub App ID. Required when using GitHub App authentication (see [GitHub App auth](#github-app-authentication)). Mutually exclusive with `token`.                                                                                                                                            |
| `githubInstallationId`    | **        | int    | **none** | GitHub App installation ID. Required when using GitHub App authentication.                                                                                                                                                                                                                    |
| `githubAppPrivateKey`     | **        | string | **none** | PEM-encoded RSA private key for the GitHub App (inline). Use this or `githubAppPrivateKeyFile`, not both.                                                                                                                                                                                     |
| `githubAppPrivateKeyFile` | **        | string | **none** | Path to a PEM file containing the GitHub App private key. Useful for Kubernetes secret volume mounts. Takes precedence over `githubAppPrivateKey` if both are set.                                                                                                                            |

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

| Field                | Mandatory | Description                                                                                                                                                                                                                                                                                 |
| -------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`RepositorySlug`** | **        | Your GitHub slug `org/repo-name`.                                                                                                                                                                                                                                                           |
| **`FilePath`**       | **        | The path of your file.                                                                                                                                                                                                                                                                      |
| **`Branch`**         | **        | The branch where your file is.<br />Default: `main`                                                                                                                                                                                                                                         |
| **`GithubToken`**    | **        | GitHub token (PAT) used to access a private repository, you need the `repo` permission *([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token))*. Mutually exclusive with the GitHub App fields. |
| **`BaseURL`**        | **        | The base URL for the GitHub API. Use this for GitHub Enterprise or self-hosted instances *(ex: `https://github.acme.com/api/v3`)*. If not specified, defaults to `https://api.github.com`.                                                                                                  |
| **`Timeout`**        | **        | Timeout for the HTTP call<br />Default: 10 seconds                                                                                                                                                                                                                                          |
| **`AppID`**          | **        | GitHub App ID. Required when using GitHub App authentication (see [GitHub App auth](#github-app-authentication)). Mutually exclusive with `GithubToken`.                                                                                                                                    |
| **`InstallationID`** | **        | GitHub App installation ID. Required when using GitHub App authentication.                                                                                                                                                                                                                  |
| **`PrivateKey`**     | **        | PEM-encoded RSA private key for the GitHub App (inline string). Use this or `PrivateKeyFile`, not both.                                                                                                                                                                                     |
| **`PrivateKeyFile`** | **        | Path to a PEM file containing the GitHub App private key. Takes precedence over `PrivateKey` if both are set.                                                                                                                                                                               |

## GitHub App authentication[​](#github-app-authentication "Direct link to GitHub App authentication")

Instead of a personal access token you can authenticate as a [GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/about-creating-github-apps). This removes the dependency on a user account and gives you fine-grained, repo-scoped permissions.

The retriever handles the full OAuth flow internally: it signs a short-lived JWT with the app's private key, exchanges it for an installation token, caches the token, and refreshes it automatically before it expires. No external service or sidecar is needed.

**GitHub App prerequisites:**

* Repository permission: **Contents: read-only** (Metadata: read is implicit).
* Install the app on the repository that holds your flag file.
* Generate a private key in the app settings and store it securely (e.g. a Kubernetes secret or AWS Secrets Manager).
* Note the **App ID** (shown on the app settings page) and **Installation ID** (visible in the URL of the app's installation settings: `https://github.com/organizations/<org>/settings/installations/<id>`).

### Relay Proxy Configuration[​](#relay-proxy-configuration "Direct link to Relay Proxy Configuration")

goff-proxy.yaml

```
# ...

retrievers:

  - kind: github

    repositorySlug: myorg/my-flags-repo

    path: flags/config.yaml

    githubAppId: 12345

    githubInstallationId: 67890

    # Option A: path to a PEM file (recommended for Kubernetes secret mounts)

    githubAppPrivateKeyFile: /var/secrets/github-app.pem

    # Option B: inline PEM string — store the actual value in a secret manager

    # and inject it via an environment variable reference in your deployment.

    # githubAppPrivateKey: <contents of your .pem file>

# ...
```

note

`token` / `githubToken` and the GitHub App fields (`githubAppId`, `githubInstallationId`, `githubAppPrivateKey` / `githubAppPrivateKeyFile`) are mutually exclusive. Validation will reject a configuration that supplies both.

### GO Module Configuration[​](#go-module-configuration "Direct link to GO Module Configuration")

example.go

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &githubretriever.Retriever{

        RepositorySlug: "myorg/my-flags-repo",

        FilePath:       "flags/config.yaml",

        AppID:          12345,

        InstallationID: 67890,

        // Option A: path to a PEM file (recommended)

        PrivateKeyFile: "/var/secrets/github-app.pem",

        // Option B: inline PEM string loaded from a secret at runtime

        // PrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),

    },

})

defer ffclient.Close()
```

## GitHub Enterprise / Self-hosted instances[​](#github-enterprise--self-hosted-instances "Direct link to GitHub Enterprise / Self-hosted instances")

If you are using a GitHub Enterprise or self-hosted instance, you can configure the `baseUrl` field to point to your instance's API endpoint.

### Relay Proxy Configuration[​](#relay-proxy-configuration-1 "Direct link to Relay Proxy Configuration")

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

### GO Module Configuration[​](#go-module-configuration-1 "Direct link to GO Module Configuration")

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
