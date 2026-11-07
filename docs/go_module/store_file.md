# 📦 Retrieve your feature flags configuration

The module supports different ways of retrieving the flag file.

To retrieve a file you need to provide a [retriever](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/#Retriever) in your `ffclient.Config{}` during the initialization.

**Available retrievers**:

* [HTTP(S)](https://gofeatureflag.org/../integrations/store-flags-configuration/http)
* [File System](https://gofeatureflag.org/../integrations/store-flags-configuration/file)
* [Kubernetes ConfigMap](https://gofeatureflag.org/../integrations/store-flags-configuration/kubernetes-configmap)
* [AWS S3](https://gofeatureflag.org/../integrations/store-flags-configuration/aws-s3)
* [MinIO](https://gofeatureflag.org/../integrations/store-flags-configuration/minio)
* [Google Cloud Storage](https://gofeatureflag.org/../integrations/store-flags-configuration/google-cloud-storage)
* [Azure Blob Storage](https://gofeatureflag.org/../integrations/store-flags-configuration/azure-blob-storage)
* [GitHub](https://gofeatureflag.org/../integrations/store-flags-configuration/github)
* [GitLab](https://gofeatureflag.org/../integrations/store-flags-configuration/gitlab)
* [Bitbucket](https://gofeatureflag.org/../integrations/store-flags-configuration/bitbucket)
* [MongoDB](https://gofeatureflag.org/../integrations/store-flags-configuration/mongodb)
* [Redis](https://gofeatureflag.org/../integrations/store-flags-configuration/redis)
* [Postgresql](https://gofeatureflag.org/../integrations/store-flags-configuration/postgresql)

info

In each retriever documentation, you will find the configuration needed to set up the retriever.

## Explicitly call the retrievers[​](#explicitly-call-the-retrievers "Direct link to Explicitly call the retrievers")

By default, the retrievers are called regularly to refresh the configuration based on the polling interval.

But there are use cases where you want to refresh the configuration explicitly *(for example, during the CI process after you have changed your configuration file)*.

To do that, you can call the `ForceRefresh` method on the client.

```
// Init ffclient with a file retriever.

	goff, _ := ffclient.New(ffclient.Config{

		PollingInterval: 10 * time.Minute,

		Retriever: &fileretriever.Retriever{

			Path: "xxxx.yaml",

		},

	})



	// ...

	goff.ForceRefresh()

	// ...
```
