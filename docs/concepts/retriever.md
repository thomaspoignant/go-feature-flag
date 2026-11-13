# 🎣 Retriever

## Overview[​](#overview "Direct link to Overview")

In GO Feature Flag, a "**retriever**" is a crucial component responsible for fetching the feature flag configuration data. It acts as the source of truth for your flags, loading the configuration from various sources like local files, remote URLs, cloud storage, or even databases.

## How it works ?[​](#how-it-works- "Direct link to How it works ?")

The retriever is responsible for loading the feature flag configuration from a specified location, it polls the source at a regular interval *(you can configure the interval)* to fetch the latest configuration. Once the configuration is loaded, it is used by the core module to evaluate the feature flags.

This simple pattern allows to decouple the feature flag configuration from the core module and to load it from various sources. This is why GO Feature Flag supports a variety of retrievers out of the box.

## Supported retrievers[​](#supported-retrievers "Direct link to Supported retrievers")

* HTTP(S)
* File System
* Kubernetes ConfigMap
* AWS S3
* MinIO
* Google Cloud Storage
* Azure Blob Storage
* GitHub
* GitLab
* Bitbucket
* MongoDB
* Redis
* Postgresql

[Check how to configure the retrievers](https://gofeatureflag.org/docs/integrations/store-flags-configuration.md)

## Using multiple retrievers[​](#using-multiple-retrievers "Direct link to Using multiple retrievers")

Sometimes, you might need to store your feature flags in different locations. In such cases, you can configure multiple retrievers to retrieve the flags from different sources within your GO Feature Flag instance.

To set this up, you need to configure the Retrievers field to consume from different retrievers. What this does is that it calls all the retrievers in parallel and applies them in the order you have provided.

note

Keep in mind that if a flag is defined in multiple retrievers, it can be overridden by a later flag.

For instance, if you have a flag named `my-feature-flag` in the first file and another flag with the same name in the second file, the second configuration will take precedence.
