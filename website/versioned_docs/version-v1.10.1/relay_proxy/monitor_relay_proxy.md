---
sidebar_position: 80
description: Monitor the relay proxy.
---

# Monitor the relay proxy

The **relay proxy** offer some endpoints for you to be able to see how it behaves.

## `/health`
Making a **GET** request to the URL path `/health` will tell you if the relay proxy is ready to
serve traffic.

This is useful especially for loadbalancer to know that they can send traffic to the service.

## `/info`
Making a **GET** request to the URL path `/info` will give you information about the actual state
of the relay proxy.

## `/metrics`
This endpoint is providing metrics about the relay proxy in the prometheus format.
