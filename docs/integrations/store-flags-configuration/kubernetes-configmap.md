# Kubernetes ConfigMap

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Loads the configuration from a Kubernetes ConfigMap. This retriever is useful when you are using Kubernetes and want to use ConfigMaps to store your configuration files.

<!-- -->

The **Kubernetes Retriever** will access flags in a Kubernetes ConfigMap via the [Kubernetes Go Client](https://github.com/kubernetes/client-go).

## Add a flag configuration file as ConfigMap[​](#add-a-flag-configuration-file-as-configmap "Direct link to Add a flag configuration file as ConfigMap")

If you have a flag configuration file, you can create a ConfigMap with the content of the file. The following command will create a ConfigMap with the content of the `examples/retriever_configmap/flags.yaml` file:

```
kubectl create configmap goff --from-file=examples/retriever_configmap/flags.yaml
```

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Kubernetes ConfigMap<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

note

Relay proxy is only supporting the configmaps while running inside the kubernetes cluster.

goff-proxy.yaml

```
# ...

retrievers:

  - kind: configmap

    namespace: default

    configmap: my-configmap

    key: my-flags.yml

# ...
```

| Field name  | Mandatory | Type   | Default  | Description                                                                                                  |
| ----------- | --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------------ |
| `kind`      | **        | string | **none** | Value should be **`configmap`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `namespace` | **        | string | **none** | This is the name of the namespace where your **configmap** is located *(ex: `default`)*.                     |
| `configmap` | **        | string | **none** | Name of the **configmap** we should read *(ex: `feature-flag`)*.                                             |
| `key`       | **        | string | **none** | Name of the `key` in the **configmap** which contains the flag.                                              |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Kubernetes ConfigMap<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
import (

    restclient "k8s.io/client-go/rest"

)

// ...

config, _ := restclient.InClusterConfig()

err = ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &k8sretriever.Retriever{

        Path: "file-example.yaml",

        Namespace:      "default"

        ConfigMapName: "my-configmap"

        Key:    "my-flags.yml"

        ClientConfig: &config

    },

})

defer ffclient.Close()
```

| Field               | Mandatory | Description                                        |
| ------------------- | --------- | -------------------------------------------------- |
| **`Namespace`**     | **        | The namespace of the ConfigMap.                    |
| **`ConfigMapName`** | **        | The name of the ConfigMap.                         |
| **`Key`**           | **        | The key within the ConfigMap storing the flags.    |
| **`ClientConfig`**  | **        | The configuration object for the Kubernetes client |
