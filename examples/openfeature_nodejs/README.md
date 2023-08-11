# GO Feature Flag Node.js integration

This example shows how to use GO Feature Flag with a simple Node.js application.

## How to start the example?

```bash
docker compose up
```

It will start `thomaspoignant/go-feature-flag` docker image and build the nodejs application located in the `nodejs-app` directory.

## What this example does?
It uses the Openfeature SDK and the GO Feature Flag provider to call the server.

The configuration of the server is in the `goff-proxy.yaml` file and it loads the flag configuration from the `config.goff.yaml` file.

In the Node.js application, we are rotating over 6 different evaluation context *(as you can see in the source)* and we check if the flag named `my-new-feature` is applying for this user.

At any moment during the demo you can edit the `config.goff.yaml` file and see how it changes the behaviors of the application.
