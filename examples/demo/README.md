# GO Feature Flag demo

This repository contains a demo app using the library [`thomaspoignant/go-feature-flag`](https://github.com/thomaspoignant/go-feature-flag) which display a webapp containing a grid of users.  
Each square is a different user and can be change by modifying the feature flag `color-box`.


With this demo app you can modify your flag and visually see which users are impacted by the change.

https://user-images.githubusercontent.com/17908063/211581747-f6354a9d-8be6-4e52-aa53-7f0d6a40827e.mp4

In this example we can see how randomly the flag apply to only a percentage of the users.


## About the app
The app use `labstack/echo` as a http server and serve an HTML page with one square per user.

Every square has his own UUID to represent a user, it means that you play with you flag and directly see which user will be impacted.

## Build the app

To build the app you have to run these command:

```shell
go mod tidy && go mod vendor # to retrieve dependencies
go build .
./demo
```

## Report a problem
If you have any issue with this demo app you can open an issue.
