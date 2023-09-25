# How to contribute to GO Feature Flag website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.  

You will need to have **nodejs** installed in your machine to work with the documentation website.


## Launch locally

Your can start locally the website.

1. Open a terminal and go to the root project of this repository.
2. Launch the command below, it will install the dependencies and run the local server for the documentation.
```shell
make watch-doc
```
3. You can now access the documentation directly in your browser: [http://localhost:3000/](http://localhost:3000/).


## Build the documentation
1. Open a terminal and go to the root project of this repository.
2. Launch the command below, it will install the dependencies and build the documentation.
```shell
make build-doc
```
3. If you want to see the result of your build, you can serve it with this command:
```shell
make serve-doc
```
