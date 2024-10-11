---
sidebar_position: 70
title: Deployment
description: Deploy the relay proxy.
---

# Deploy the relay proxy

## Deploy in Kubernetes using Helm

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

Install the Helm Chart with the Helm install command and provide the custom repository name, the chart name and any
necessary values files.  
You can look at
the [helm doc](https://github.com/thomaspoignant/go-feature-flag/blob/main/cmd/relayproxy/helm-charts/relay-proxy/README.md)
to know exactly what you can change in the values.yaml file.

```shell
helm install go-feature-flag/relay-proxy -f values.yaml
```

### Step 3: Verify The Chart Installation

Verify the Helm Chart installation with the Helm list command. For example:

```shell
helm list
```

## Deploy as AWS Lambda

The GO Feature Flag relay proxy can easily be launched as an AWS Lambda function.
To do this, simply set the `startAsAwsLambda` option in your configuration file to `true`, like so:

```yaml
# ...
startAsAwsLambda: true
```

Once you've updated your configuration file, you can deploy your function in AWS and configure it to be accessible
via HTTP. This can be achieved by creating an API Gateway or an Application Load Balancer (ALB) and linking it to
your Lambda function.

By configuring your GO Feature Flag relay proxy to run as an AWS Lambda function, you can take advantage of many
benefits of serverless computing, including automatic scaling, reduced infrastructure costs, and simplified
deployment and management.

:::info
As part of our release process, we are building an archive ready to be deployed as AWS lambda.  
You can find it in the [GitHub release page](https://github.com/thomaspoignant/go-feature-flag/releases),and you can use
the assets named `go-feature-flag-aws-lambda_<version>.zip`.
:::

### Choose the handler for your AWS Lambda

Depending on what you put in front of your Lambda function, you will need to choose the right handler.    
GO Feature Flag supports 3 different handlers:

- `APIGatewayV1`: This handler is used when you put an API Gateway with the v1 format in front of your Lambda function.
- `APIGatewayV2`: This handler is used when you put an API Gateway with the v2 format in front of your Lambda function.
- `ALB`: This handler is used when you put an Application Load Balancer in front of your Lambda function.

To choose the right handler, you need to set the `awsLambdaAdapter` option in your configuration file with one of this
value. If you don't set this option, the default value is `APIGatewayV2`.

```yaml
# ...
startAsAwsLambda: true
awsLambdaAdapter: ALB
```