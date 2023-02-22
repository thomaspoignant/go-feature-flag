# `relay-proxy` Helm Chart

This folder contains a basic helm-chart to deploy `go-feature-flag-relay-proxy` inside Kubernetes.

Please [create an issue](https://github.com/thomaspoignant/go-feature-flag/issues/new/choose) or submit a pull request
for any issues or missing features.


## How to use the chart.
Please replace the file `goff-proxy.yml` in this folder by your `relay-proxy` configuration file.  
This file will be stored as `configmap` in your cluster and be mount as a volume for the `relay-proxy`.

After in the current folder run:
```shell
helm install . --name-template=go-feature-flag-relay-proxy
```

It will install the chart in your cluster.

## Advanced
You can edit the `values.yaml` file to enable an ingress or the autoscaling.