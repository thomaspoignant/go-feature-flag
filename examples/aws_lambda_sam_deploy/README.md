# Using AWS SAM to Test the Relay-Proxy as an AWS Lambda

This README provides instructions on how to use AWS SAM to test the relay-proxy as an AWS Lambda.

## Prerequisites

Before starting, you need to have AWS SAM installed. If you don't have AWS SAM installed, follow these steps:

1. Download and install the AWS SAM CLI by following the instructions in the [AWS SAM CLI Installation Guide](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html).
2. Verify that AWS SAM is installed correctly by running the following command in your terminal: `sam --version`

## Usage

Once you have AWS SAM installed, you can use it to test the relay-proxy as an AWS Lambda. Here are the steps to follow:

1. Clone this repository to your local machine.
2. Navigate to this example directory in your terminal (`/examples/aws_lambda_deploy/`).
3. Run the following command to build the Lambda function: `make build`
4. Start the Lambda function locally by running the following command: `make start-api`
5. Call the API by running the following command: `make call-api`

The `make build` command compiles the Lambda function using the AWS SAM CLI.  
The `make start-api` command starts the Lambda function locally, allowing you to test it without deploying it to AWS.
Finally, in another terminal the `make call-api` command sends a request to the API to test it.

That's it! You can now test the `relay-proxy` as an AWS Lambda using AWS SAM.

## Conclusion

In this README, you learned how to use AWS SAM to test the relay-proxy as an AWS Lambda.
With these instructions, you can easily test your Lambda function locally before deploying it to AWS.

:warning: This example is not production ready.