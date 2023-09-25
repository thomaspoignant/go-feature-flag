---
sidebar_position: 6
---

# GitLab

The [**Gitlab Retriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/gitlabretriever/#Retriever)
will perform an HTTP Request to the Gitlab API to get your flags.

!!! Tip
    Gitlab has rate limits, be sure to correctly set your `PollingInterval` to avoid reaching the limit.

## Example

```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &gitlab.Retriever{
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

## Configuration fields

To configure the access to your GitLab file:

| Field                | Description                                                                               |
|----------------------|-------------------------------------------------------------------------------------------|
| **`BaseURL`**        | *(optional)*<br/>The domain name of your Gitlab instance <br/>Default: https://gitlab.com |
| **`RepositorySlug`** | Your Gitlab slug `org/repo-name`.                                                         |
| **`FilePath`**       | The path of your file.                                                                    |
| **`Branch`**         | *(optional)*<br/>The branch where your file is.<br/>Default: `main`                       |
| **`GitlabToken`**    | *(optional)*<br/>Gitlab token is used to access a private repository                      |
| **`Timeout`**        | *(optional)*<br/>Timeout for the HTTP call <br/>Default: 10 seconds                       |
