---
sidebar_position: 10
description: How to deploy GO Feature Flag relay proxy with Helm.
---

# Helm

## Overview
The relay proxy can be deployed in Kubernetes using a helm chart.  
Helm is an invaluable tool for configuring and deploying applications to a Kubernetes environment.

We provide a Helm chart to deploy the relay proxy in your Kubernetes cluster and it is available in our Helm repository (https://charts.gofeatureflag.org/).

Below are the steps for installing a Helm Chart from a **GO Feature Flag** Helm repository.

## Install GO Feature Flag relay proxy in Kubernetes using Helm
### Prerequisites

- Access to a Kubernetes cluster
- Helm CLI installed on the client machine

### Step 1: Prepare and Configure the Repository

Add the repository to Helm with the Helm repository add command and provide a name and the repository URL. For example:

```shell
helm repo add go-feature-flag https://charts.gofeatureflag.org/
```

### Step 2: Install the Chart

Install the Helm Chart with the Helm install command and provide the custom repository name, the chart name and any necessary values files.  
You can look at the [helm doc](https://github.com/thomaspoignant/go-feature-flag/blob/main/cmd/relayproxy/helm-charts/relay-proxy/README.md) to know exactly what you can change in the values.yaml file.

```shell
helm install go-feature-flag/relay-proxy -f values.yaml
```

### Step 3: Verify The Chart Installation

Verify the Helm Chart installation with the Helm list command. For example:

```shell
helm list
```

## Using the GitHub Container Registry

If your infrastructure cannot reach Docker Hub, both the chart and the relay proxy image
are also available on the GitHub Container Registry.

### Installing the chart from GHCR

The chart is published as an OCI artifact, so there is no repository to add:

```shell
helm install relay-proxy oci://ghcr.io/go-feature-flag/charts/relay-proxy --version <chart-version> -f values.yaml
```

### Pulling the image from GHCR

The chart still defaults to the Docker Hub image, override `image.repository` to pull it
from GHCR instead:

```yaml title="values.yaml"
image:
  repository: ghcr.io/go-feature-flag/go-feature-flag
```

:::warning
Images are mirrored to GHCR only for the versions released after `v1.55.1`.

Since the chart defaults `image.tag` to its own `appVersion`, this override requires a
chart published **after** the mirroring started. Chart `1.55.1` pins `appVersion v1.55.1`,
which predates the mirror and has no GHCR image, so combining it with this override fails
with `ImagePullBackOff`. The same applies if you pin an older `image.tag` explicitly.
:::
