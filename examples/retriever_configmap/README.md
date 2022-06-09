# Kubernetes config map example

This example contains everything you need to use a **`configmap`** as the source for your flags.  
We will use minikube to test the solution, but it works the same in your cluster.

As you can see the `main.go` file contains a basic HTTP server that expose an API that use your flags.  
For this example we are using a `InClusterConfig` because we will run the service inside kubernetes.

## How to setup the example
_All commands should be run in the root level of the repository._

1. Load all dependencies

```shell
make vendor
```

2. Create a minikube environment in your machine:

```shell
minikube start --vm
```

3. Use the minikube docker cli in your shell

```shell
eval $(minikube docker-env)
```

4. Build the docker image of the service

```shell
docker build -f examples/retriever_configmap/Dockerfile -t goff-test-configmap .
```

5. Create a `configmap` based on your `go-feature-flag` config file

```shell
kubectl create configmap goff --from-file=examples/retriever_configmap/flags.yaml
```

6. Deploy your service to your kubernetes instance

```shell
kubectl apply -f examples/retriever_configmap/k8s-manifests.yaml
```

7. Forward the port to the service

```shell
kubectl port-forward $(kubectl get pod | grep "goff-test-configmap" | cut -d ' ' -f1) 9090:8080
```

8. Access to the service and check the values for different users

```shell
curl http://localhost:9090/
```

9. Play with the values in the `go-feature-flag` config file

```shell
kubectl edit configmap goff
```

10. Delete your minikube instance

```shell
minikube delete
```
