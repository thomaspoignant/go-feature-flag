# Contributing

When contributing to this repository, please first discuss the change you wish to make via an issue.  
Please note we have a [code of conduct](CODE_OF_CONDUCT.md), please follow it in all your interactions with the project.

# Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2. Please mention the issue in your PR description.
3. Expect to be taken seriously, if there are some feedbacks, feel free to discuss about it, your opinion can be better than mine.

# Coding standards

A library is easier to use, and easier for contributors to work on if it has a consistent, unified style, approach, and layout.

We are using [pre-commit](https://pre-commit.com/) to lint before each commit, I would recommend you to use it.
```bash
pre-commit install
```

## Tests

Every feature or bug should come with an associate test to keep the coverage as high as possible.

## Documentation

We are maintaining 2 documentations:
- [README.md](README.md) which contains everything you need to know to start working with the module.
- [go-feature-flag website](https://thomaspoignant.github.io/go-feature-flag/) which is the full detail documentation of the module.

If your contribution has impact on the documentation, please check both version.

### How to run the documentation website locally

For the documentation website we are using [mkdocs](https://www.mkdocs.org/) with the "[Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)" theme.
To run it locally just use the docker image:
```shell
docker run --rm -it -p 8000:8000 -v ${PWD}:/docs squidfunk/mkdocs-material
```
The website will be available on http://localhost:8000/

## Sonar

Sonarcloud is used in the project, it will comment your PR to give you feedback on your code.

### Continuous integration

We have a list of steps on each PR.  
The CI is running:

 - Tests
 - Coverage
 - Code quality

With this CI you will have feedbacks on your PR after opening your PR. Please review it if it fails.
