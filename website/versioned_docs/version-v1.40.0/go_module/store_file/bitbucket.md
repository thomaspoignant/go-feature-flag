---
sidebar_position: 7
---

# Bitbucket

The [**Bitbucket Retriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/bitbucketretriever/#Retriever)
will perform an HTTP Request to the Bitbucket API to get your flags.

:::tip
Bitbucket has rate limits, be sure to correctly set your `PollingInterval` to avoid reaching the limit.
:::

## Example

```go showLineNumbers
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

## Configuration fields

To configure the access to your GitLab file:

| Field                | Description                                                                                           |
|----------------------|-------------------------------------------------------------------------------------------------------|
| **`BaseURL`**        | *(optional)*<br/>The domain name of your Bitbucket instance <br/>Default: `https://api.bitbucket.org` |
| **`RepositorySlug`** | Your Gitlab slug `org/repo-name`.                                                                     |
| **`FilePath`**       | The path of your file.                                                                                |
| **`Branch`**         | *(optional)*<br/>The branch where your file is.<br/>Default: `main`                                   |
| **`BitBucketToken`** | *(optional)*<br/>Bitbucket token is used to access a private repository                               |
| **`Timeout`**        | *(optional)*<br/>Timeout for the HTTP call <br/>Default: 10 seconds                                   |
