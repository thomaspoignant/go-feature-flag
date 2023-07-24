# Contributing to GO Feature Flag

Thank you for considering contributing to this repository!  
You will contribute to the coolest Opensource Feature Flag solution and be part of an awesome community 😜.

We welcome contributions to improve and grow this project.  
Please take a moment to review the following guidelines.

## Table of Contents

1. [Code of Conduct](#-code-of-conduct)
2. [How Can I Contribute?](#-how-can-i-contribute)
3. [Development Setup](#-development-setup)
4. [Documentation](#-documentation)
5. [License](#-license)


## 🚓 Code of Conduct
We expect all contributors to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).  
Please read it thoroughly before contributing.

## 🙇 How Can I Contribute?

### Bug Reports
If you encounter any bugs or issues with the project, please [create a new issue](../../issues/new?assignees=&labels=bug%2Cneeds-triage&projects=&template=bug.yaml&title=%28bug%29+%3Ctitle%3E) and include as many details as possible, such as:

- A clear and descriptive title
- Steps to reproduce the bug
- Expected behavior and actual behavior
- Your operating system and version
- The version of the project you were using when the bug occurred
- Any relevant error messages or logs

### Feature Requests

If you have a feature idea that you would like to see implemented, please [create a new issue](../../issues/new?assignees=&labels=enhancement%2Cneeds-triage&projects=&template=feature.yaml&title=(feature)+<title>) with the following information:

- A clear and descriptive title
- A detailed description of the feature
- Any additional context that may be relevant

### Pull Requests

We welcome contributions in the form of pull requests. 

Before opening a pull request, we kindly request you to check if there is an open issue related to your proposed contribution.
By doing so, we can initiate a discussion and provide feedback on your changes before proceeding with the pull request.
This approach ensures that your efforts align with the project's goals and enhances the chances of your contribution being successfully integrated. Thank you for your understanding and cooperation!

To submit a pull request, follow these steps:

1. Fork the repository to your GitHub account.
2. Create a new branch from the `main` branch.
3. Make your changes.
4. Test your changes thoroughly.
5. Commit your changes with a clear and descriptive commit message.
6. Push your branch to your forked repository.
7. Open a pull request, comparing your branch to the `main` branch of this repository.
8. Provide a detailed description of your changes in the pull request.

We will review your pull request as soon as possible
Please be patient, as it might take some time for us to get back to you
Your contributions are highly valued!

## 🧑‍💻 Development Setup
We always thrive to keep the project as simple as possible, so you will find everything you need in the `Makefile` at the root of the repository.

To start contributing please set up your GO environment and run: 

```shell
make vendor
```
It will download the dependencies and your project will be ready to be used.

### Coding standards

It is easier for contributors to work on the same project if it has a consistent, unified style, approach, and layout.

To help with that, we are using [pre-commit](https://pre-commit.com/) to lint before each commit, I would recommend you to install it, and to apply it to the project by running:
```bash
pre-commit install
```

### Tests
Every feature or bug should come with an associate test to keep the coverage as high as possible.
We aim to have 90% of coverage for the project.

## 📚 Documentation

We are maintaining 2 documentations:
- [README.md](README.md) which contains everything you need to know to start working with the module.
- [go-feature-flag website](https://gofeatureflag.org) which is the full detail website containing the documentation.

If your contribution has impact on the documentation, please check both version. You can check how to work on the documentation [here](./website/README.md).

### How to run the documentation website locally

For the documentation website we are using [Docusaurus 2](https://docusaurus.io/).  
Everything is available in the [`website/docs`](website/docs) directory.

Your can start locally the website.

1. Open a terminal and go to the root project of this repository.
2. Launch the command bellow, it will install the dependencies and run the local server for the documentation.
```shell
make watch-doc
```
3. You can now access to the documentation directly in your browser: [http://localhost:3000/](http://localhost:3000/).

## 🪪 License

By contributing to this repository, you agree that your contributions will be licensed under the [LICENSE](LICENSE) of the project.

---

We encourage everyone to participate in this project and make it better for everyone. Happy contributing 🎉 
