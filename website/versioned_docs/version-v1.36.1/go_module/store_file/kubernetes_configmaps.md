---
sidebar_position: 4
---

# Kubernetes configmaps
A ConfigMap is an API object used to store non-confidential data in key-value pairs inside kubernetes.  
GO Feature Flag can read directly in a configmap in your namespace.

The [**Kubernetes Retriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/k8sretriever/#Retriever)
will access flags in a Kubernetes ConfigMap via the [Kubernetes Go Client](https://github.com/kubernetes/client-go).


## Add your config file as configmap

```shell
kubectl create configmap goff --from-file=examples/retriever_configmap/flags.yaml
```

## Configuration Example
```go showLineNumbers
import (
    restclient "k8s.io/client-go/rest"
)

config, _ := restclient.InClusterConfig()
err = ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &k8sretriever.Retriever{
        Path: "file-example.yaml",
        Namespace:      "default"
        ConfigMapName: "my-configmap"
        Key:    "somekey.yml"
        ClientConfig: &config
    },
})
defer ffclient.Close()
```

## Configuration fields
To configure your retriever:

| Field               | Description                                        |
|---------------------|----------------------------------------------------|
| **`Namespace`**     | The namespace of the ConfigMap.                    |
| **`ConfigMapName`** | The name of the ConfigMap.                         |
| **`Key`**           | The key within the ConfigMap storing the flags.    |
| **`ClientConfig`**  | The configuration object for the Kubernetes client |
