---
sidebar_position: 10
description: Where to store your configuration flag?
---

# Where to store your configuration flag

To work **GO Feature Flag** is using a configuration file to store your feature flags configuration.

Ideally this file should be placed somewhere accessible by all your app that are using the GO Feature Flag module.  
In the solution there is a system we call `retriever` that is in charge of reading the file remotely.

GO Feature Flag supports different ways to retrieve the configuration file _(see list bellow)_.

## Use multiple configuration flag files
Sometimes, you may need to store your feature flags in different locations.
In such cases, you can configure multiple retrievers to retrieve the flags from different sources within your GO Feature
Flag instance.

To set this up, you need to configure the [`Retrievers`](../go_module/configuration#configuration-fields) field to 
consume from different retrievers.
What this does is that it calls all the retrievers in parallel and applies them in the order you have provided.

Keep in mind that if a flag is defined in multiple retrievers, it can be overridden by a later flag. For instance, 
if you have a flag named _`my-feature-flag`_ in the first file and another flag with the same name in the second file, the second configuration will take precedence.

## Available retrievers
### AWS S3

**AWS S3** is the object store of AWS, you can add your GO Feature Flag configuration file in any bucket and reference it in your configuration.

- [Configure the GO module](../go_module/store_file/s3.md) 
- [Configure the relay proxy](../relay_proxy/configure_relay_proxy#s3)

### Google Cloud Storage

**Google Cloud Storage** is a RESTful online file storage web service for storing and accessing data on Google Cloud Platform infrastructure.  
You can add your GO Feature Flag configuration file in any bucket and reference it.

- [Configure the GO module](../go_module/store_file/google_cloud_storage.md) 
- [Configure the relay proxy](../relay_proxy/configure_relay_proxy#google-storage)

### Kubernetes Configmaps

A **ConfigMap** is an API object used to store non-confidential data in key-value pairs inside kubernetes.  
GO Feature Flag can read directly in a `configmap` in your namespace.

When your feature flags file is ready you can store it directly in your kubernetes instance by using this command:

```shell
kubectl create configmap goff --from-file=examples/retriever_configmap/flags.yaml
```

It will allow your file to be available inside Kubernetes.

- [Configure the GO module](../go_module/store_file/kubernetes_configmaps.md) 
- [Configure the relay proxy](../relay_proxy/configure_relay_proxy#kubernetes-configmap)

### HTTP / HTTPS

Serving file with an **HTTP** server is probably something you are already doing, **GO Feature Flag** can retrieve your configuration file, from 
any HTTP endpoint and read it from there.

- [Configure the GO module](../go_module/store_file/http.md) 
- [Configure the relay proxy](../relay_proxy/configure_relay_proxy#http)

### GitHub

Reading the file from **GitHub** is pretty straight forward.  
You commit your file into your favorite repository _(it can be public or private)_ and **GO Feature Flag** can retrieve your configuration file and use it.

- [Configure the GO module](../go_module/store_file/github.md) 
- [Configure the relay proxy](../relay_proxy/configure_relay_proxy#github)

### Local file

You can store your feature flags configuration in your hard drive directly.

:::tip
Using a file is great for local testing, but in production it is recommended to use a distributed system instead.
:::

- [Configure the GO module](../go_module/store_file/file.md) 
- [Configure the relay proxy](../relay_proxy/configure_relay_proxy#file)

### Custom

If you are using the **GO module**, you can also implement your own retriever.  
For this look at this [documentation](../go_module/store_file/custom.md) to start building your own `retriever`.
