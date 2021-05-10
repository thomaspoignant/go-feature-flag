# Store your flag file
`go-feature-flags` support different ways of retrieving the flag file.  
We can have only one source for the file, if you set multiple sources in your configuration, only one will be take in
consideration.

Available retriever are:

- [Github](#from-github)
- [HTTP endpoint](#from-an-http-endpoint)
- [S3 Bucket](#from-a-s3-bucket)
- [File](#from-a-file)

### From Github

```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.GithubRetriever{
        RepositorySlug: "thomaspoignant/go-feature-flag",
        Branch: "main",
        FilePath: "testdata/flag-config.yaml",
        GithubToken: "XXXX",
        Timeout: 2 * time.Second,
    },
})
defer ffclient.Close()
```
To configure the access to your GitHub file:

| Field | Description |
|---|---|
|**`RepositorySlug`**| your GitHub slug `org/repo-name`.|
|**`FilePath`**| the path of your file.|
|**`Branch`**| *(optional)* the branch where your file is *(default is `main`)*.|
|**`GithubToken`**| *(optional)* Github token is used to access a private repository, you need the `repo` permission *([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token))*.|
|**`Timeout`**| *(optional)* Timeout for the HTTP call <br>(default is 10 seconds).|

!!! warning
    GitHub has rate limits, so be sure to not reach them when setting your `PollInterval`.

### From an HTTP endpoint

```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/flag-config.yaml",
        Timeout: 2 * time.Second,
    },
})
defer ffclient.Close()
```

To configure your HTTP endpoint:

| Field | Description |
|---|---|
|**`URL`**| location of your file.|
|**`Method`**| the HTTP method you want to use <br>*(default is GET)*.|
|**`Body`**| *(optional)* If you need a body to get the flags.|
|**`Header`**| *(optional)* Header you should pass while calling the endpoint *(useful for authorization)*.|
|**`Timeout`**| *(optional)* Timeout for the HTTP call <br>(default is 10 seconds).|

### From a S3 Bucket

```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.S3Retriever{
        Bucket: "tpoi-test",
        Item:   "flag-config.yaml",
        AwsConfig: aws.Config{
            Region: aws.String("eu-west-1"),
        },
    },
})
defer ffclient.Close()
```

To configure your S3 file location:

| Field | Description |
|---|---|
|**`Bucket`**| The name of your bucket.|
|**`Item`**| The location of your file in the bucket.|
|**`AwsConfig`**| An instance of `aws.Config` that configure your access to AWS <br>*check [this documentation for more info](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)*.|

### From a file
!!! tip
    I will not recommend using a file to store your flags except if it is in a shared folder for all your services.

```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.FileRetriever{
        Path: "file-example.yaml",
    },
})
defer ffclient.Close()
```

To configure your File retriever:

| Field | Description |
|---|---|
|**`Path`**| location of your file on the file system.|


