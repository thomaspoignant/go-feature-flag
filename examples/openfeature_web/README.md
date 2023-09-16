# GO Feature Flag web integration
This example shows how to use GO Feature Flag in your web application using the Openfeature web SDK and the GO Feature Flag web provider.

## How to start the example?

```bash
docker compose up
```

It will start `thomaspoignant/go-feature-flag` docker image and build the web application located in the `webapp` directory.

When ready you can access to the application at http://localhost:3000/.

## What this example does?
It uses the Openfeature Web SDK and the GO Feature Flag web provider.

The configuration of the server is in the `goff-proxy.yaml` file and it loads the flag configuration from the `config.goff.yaml` file.

You can look at the file [`src/js/main.js`](src/js/main.js) to look how we retrieve the flags and we change the display of the page.

At any moment during the demo you can edit the `config.goff.yaml` file and see how it changes the behaviors of the application.
