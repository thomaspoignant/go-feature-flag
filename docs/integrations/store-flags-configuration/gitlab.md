# GitLab

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Fetch the configuration from files stored in a Gitlab repository. This retriever will perform an HTTP Request with your Gitlab configuration on the Gitlab API to get your flags.

<!-- -->

tip

GitLab has rate limits, be sure to correctly set your `PollingInterval` to avoid reaching the limit.

note

The retriever handles URL encoding, so write the repository slug and the file path as they appear in GitLab:

* `repositorySlug` / `RepositorySlug`: `group/subgroup/project`, not `group%2Fsubgroup%2Fproject`.
* `path` / `FilePath`: `flag configs/my flags.yaml`, without replacing spaces with `%20` or `+`.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->GitLab<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: gitlab

    repositorySlug: thomaspoignant/go-feature-flag

    path: config/flag/my-flags.yaml

# ...
```

| Field name       | Mandatory | Type   | Default              | Description                                                                                                                                                                                              |
| ---------------- | --------- | ------ | -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`           | **        | string | **none**             | **Value should be `gitlab`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                |
| `repositorySlug` | **        | string | **none**             | The repository slug of the GitLab repository where your file is located *(ex: `thomaspoignant/go-feature-flag`)*.                                                                                        |
| `path`           | **        | string | **none**             | Path to the file inside the repository *(ex: `config/flag/my-flags.yaml`)*.                                                                                                                              |
| `baseUrl`        | **        | string | `https://gitlab.com` | The base URL of your GitLab instance.                                                                                                                                                                    |
| `branch`         | **        | string | `main`               | The branch we should check in the repository.                                                                                                                                                            |
| `token`          | **        | string | **none**             | GitLab personal access token used to access a private repository ([Create a personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token)). |
| `timeout`        | **        | string | `10000`              | Timeout in millisecond used when calling GitLab.                                                                                                                                                         |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->GitLab<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &gitlabretriever.Retriever{

        RepositorySlug: "thomaspoignant/go-feature-flag",

        Branch: "main",

        FilePath: "testdata/flag-config.goff.yaml",

        GitlabToken: "XXXX",

        Timeout: 2 * time.Second,

		    BaseURL: "https://gitlab.com",

    },

})

defer ffclient.Close()
```

| Field                | Mandatory | Description                                                                |
| -------------------- | --------- | -------------------------------------------------------------------------- |
| **`BaseURL`**        | **        | The domain name of your Gitlab instance<br />Default: `https://gitlab.com` |
| **`RepositorySlug`** | **        | Your Gitlab slug `org/repo-name`.                                          |
| **`FilePath`**       | **        | The path of your file.                                                     |
| **`Branch`**         | **        | The branch where your file is.<br />Default: `main`                        |
| **`GitlabToken`**    | **        | GitLab token is used to access a private repository                        |
| **`Timeout`**        | **        | Timeout for the HTTP call<br />Default: 10 seconds                         |
