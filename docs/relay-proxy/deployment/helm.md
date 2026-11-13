# Helm

## Overview[​](#overview "Direct link to Overview")

The relay proxy can be deployed in Kubernetes using a helm chart.<br /><!-- -->Helm is an invaluable tool for configuring and deploying applications to a Kubernetes environment.

We provide a Helm chart to deploy the relay proxy in your Kubernetes cluster and it is available in our Helm repository (<https://charts.gofeatureflag.org/>).

The chart is released together with the relay proxy: its version is always the version of the relay proxy it deploys *(chart `1.55.2` deploys relay proxy `v1.55.2`)*, and a chart version is never republished once released.

Below are the steps for installing a Helm Chart from a **GO Feature Flag** Helm repository.

## Install GO Feature Flag relay proxy in Kubernetes using Helm[​](#install-go-feature-flag-relay-proxy-in-kubernetes-using-helm "Direct link to Install GO Feature Flag relay proxy in Kubernetes using Helm")

### Prerequisites[​](#prerequisites "Direct link to Prerequisites")

* Access to a Kubernetes cluster
* Helm CLI installed on the client machine

### Step 1: Prepare and Configure the Repository[​](#step-1-prepare-and-configure-the-repository "Direct link to Step 1: Prepare and Configure the Repository")

Add the repository to Helm with the Helm repository add command and provide a name and the repository URL. For example:

```
helm repo add go-feature-flag https://charts.gofeatureflag.org/
```

### Step 2: Install the Chart[​](#step-2-install-the-chart "Direct link to Step 2: Install the Chart")

Install the Helm Chart with the Helm install command and provide the custom repository name, the chart name and any necessary values files.<br /><!-- -->You can look at the [helm doc](https://github.com/thomaspoignant/go-feature-flag/blob/main/cmd/relayproxy/helm-charts/relay-proxy/README.md) to know exactly what you can change in the values.yaml file.

```
helm install go-feature-flag/relay-proxy -f values.yaml
```

### Step 3: Verify The Chart Installation[​](#step-3-verify-the-chart-installation "Direct link to Step 3: Verify The Chart Installation")

Verify the Helm Chart installation with the Helm list command. For example:

```
helm list
```

## Using the GitHub Container Registry[​](#using-the-github-container-registry "Direct link to Using the GitHub Container Registry")

If your infrastructure cannot reach Docker Hub, both the chart and the relay proxy image are also available on the GitHub Container Registry.

### Installing the chart from GHCR[​](#installing-the-chart-from-ghcr "Direct link to Installing the chart from GHCR")

The chart is published as an OCI artifact, so there is no repository to add:

```
helm install relay-proxy oci://ghcr.io/go-feature-flag/charts/relay-proxy --version <chart-version> -f values.yaml
```

### Pulling the image from GHCR[​](#pulling-the-image-from-ghcr "Direct link to Pulling the image from GHCR")

The chart still defaults to the Docker Hub image, override `image.repository` to pull it from GHCR instead:

values.yaml

```
image:

  repository: ghcr.io/go-feature-flag/go-feature-flag
```

warning

Images are mirrored to GHCR only for the versions released after `v1.55.1`.

Since the chart defaults `image.tag` to its own `appVersion`, this override requires a chart published **after** the mirroring started. Chart `1.55.1` pins `appVersion v1.55.1`, which predates the mirror and has no GHCR image, so combining it with this override fails with `ImagePullBackOff`. The same applies if you pin an older `image.tag` explicitly.
