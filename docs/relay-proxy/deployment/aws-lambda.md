# AWS Lambda

## Overview[​](#overview "Direct link to Overview")

The GO Feature Flag relay proxy can easily be launched as an AWS Lambda function.

## How to start the relay proxy as an AWS Lambda function[​](#how-to-start-the-relay-proxy-as-an-aws-lambda-function "Direct link to How to start the relay proxy as an AWS Lambda function")

To start GO Feature Flag relay-proxy as a lambda you have to set the option `startAsAwsLambda` in your configuration file to `true`, and you have to set which lambda handler you want to use, by setting the option `awsLambdaAdapter`.

goff-proxy.yaml

```
# ...

startAsAwsLambda: true

awsLambdaAdapter: APIGatewayV2

# ...
```

## API Gateway Non-Root Path Configuration[​](#api-gateway-non-root-path-configuration "Direct link to API Gateway Non-Root Path Configuration")

If you're deploying your Lambda function behind AWS API Gateway with a non-root path (e.g., `/api/feature-flags/{proxy+}`), you'll need to configure the base path to ensure proper request routing:

goff-proxy.yaml

```
# ...

startAsAwsLambda: true

awsLambdaAdapter: APIGatewayV2

awsApiGatewayBasePath: "/api/feature-flags"  # Strip this prefix from all requests

# ...
```

**Example scenario:**

* Your API Gateway route: `/api/feature-flags/{proxy+}`
* Client request: `https://your-api.amazonaws.com/api/feature-flags/health`
* With `awsApiGatewayBasePath: "/api/feature-flags"` configured, the request will be processed as `/health`

This configuration is essential when using API Gateway stages or organizing your APIs with path-based routing.

| Field name                | Mandatory | Type      | Default            | Description                                                                                                                                                                                                                                                                                                      |
| ------------------------- | --------- | --------- | ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **startAsAwsLambda**      | **        | `boolean` | **`false`**        | Set to `true` if you want to start the GO Feature Flag relay proxy as an AWS Lambda                                                                                                                                                                                                                              |
| **awsLambdaAdapter**      | **        | `string`  | **`APIGatewayV2`** | This param is used only if `startAsAwsLambda` is true.This parameter allow you to decide which type of AWS lambda handler you wan to use.Accepted values are:- `APIGatewayV2`<br />- `APIGatewayV1`<br />- `ALB`                                                                                                 |
| **awsApiGatewayBasePath** | **        | `string`  | **none**           | This param is used only if `startAsAwsLambda` is true.Specifies the base path prefix for AWS API Gateway deployments when using non-root routes. The relay proxy will strip this base path from incoming requests before processing them.Useful when deploying behind paths like `/api` or `/dev/feature-flags`. |

Once you've updated your configuration file, you can deploy your function in AWS and configure it to be accessible via HTTP. This can be achieved by creating an API Gateway or an Application Load Balancer (ALB) and linking it to your Lambda function.

By configuring your GO Feature Flag relay proxy to run as an AWS Lambda function, you can take advantage of many benefits of serverless computing, including automatic scaling, reduced infrastructure costs, and simplified deployment and management.

## Lambda Binary[​](#lambda-binary "Direct link to Lambda Binary")

As part of our release process, we are building an archive ready to be deployed as AWS lambda.

You can find it in the [GitHub release page](https://github.com/thomaspoignant/go-feature-flag/releases),and you can use the assets named `go-feature-flag-aws-lambda_<version>.zip`.
