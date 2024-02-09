---
sidebar_position: 1
---

# Retrieve your feature flags configuration
The module supports different ways of retrieving the flag file.  
Available retrievers are:

- [S3 Bucket](./s3.md)
- [HTTP endpoint](./http.md)
- [Github](./github.md)
- [Gitlab](./gitlab.md)
- [File](./file.md)
- [Kubernetes configmap](./kubernetes_configmaps.md)
- [Google Cloud storage](./google_cloud_storage.md)

To retrieve a file you need to provide a [retriever](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/#Retriever) in your `ffclient.Config{}` during the initialization.  
If the existing retriever does not work with your system you can extend the system and use a [custom retriever](custom.md).
