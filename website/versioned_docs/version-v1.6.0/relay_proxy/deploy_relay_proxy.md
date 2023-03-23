---
sidebar_position: 70
description: Deploy the relay proxy.
---

# Deploy the relay proxy

##  Deploy in Kubernetes using Helm
The relay proxy can be deployed in Kubernetes using a helm chart.  
Helm is an invaluable tool for configuring and deploying applications to a Kubernetes environment.

Below are the steps for installing a Helm Chart from a **GO Feature Flag** Helm repository.

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

```shell
helm install go-feature-flag/relay-proxy -f values.yaml
```

### Step 3: Verify The Chart Installation

Verify the Helm Chart installation with the Helm list command. For example:

```shell
helm list
```