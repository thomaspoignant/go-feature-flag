# Changelog

## [0.5.0](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.4.1...modules/core/v0.5.0) (2026-01-30)


### 🚀 New Features

* add compile-time interface checks ([#4699](https://github.com/thomaspoignant/go-feature-flag/issues/4699)) ([e4849d1](https://github.com/thomaspoignant/go-feature-flag/commit/e4849d196afc07b2a0466ccec55077ea4f2d1b64))
* **core:** merge Disable field in MergeRules for scheduled rollout ([#4726](https://github.com/thomaspoignant/go-feature-flag/issues/4726)) ([53c1b11](https://github.com/thomaspoignant/go-feature-flag/commit/53c1b11c4bfb158445fc73bce938d55460841120))


### 🐛 Bug Fixes

* error when JsonLogic panics ([#4713](https://github.com/thomaspoignant/go-feature-flag/issues/4713)) ([ac70a0a](https://github.com/thomaspoignant/go-feature-flag/commit/ac70a0ab693cbeb337c90e8cc81316605b3219b8))

## [0.4.1](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.4.0...modules/core/v0.4.1) (2026-01-20)


### 📚 Documentation

* **modules/core:** add migration notice to evaluation package ([#4667](https://github.com/thomaspoignant/go-feature-flag/issues/4667)) ([6bee4d3](https://github.com/thomaspoignant/go-feature-flag/commit/6bee4d3cdbc5c0bfa1bc9d94195291b244fbba7f))

## [0.4.0](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.3.1...modules/core/v0.4.0) (2026-01-19)


### 🚀 New Features

* Add cap to progressive rollout percentage ([#4634](https://github.com/thomaspoignant/go-feature-flag/issues/4634)) ([23c0fa4](https://github.com/thomaspoignant/go-feature-flag/commit/23c0fa4bb72e1d0c9785d709d74e762a40bd263d))


### 🐛 Bug Fixes

* **core:** include targetingKey in context map ([#4657](https://github.com/thomaspoignant/go-feature-flag/issues/4657)) ([9395a7a](https://github.com/thomaspoignant/go-feature-flag/commit/9395a7a448749b832c0fcc0524244ed554886137))


### 🔧 Chores

* Bump github.com/aws/aws-sdk-go-v2/service/s3 ([#4613](https://github.com/thomaspoignant/go-feature-flag/issues/4613)) ([a8e9b13](https://github.com/thomaspoignant/go-feature-flag/commit/a8e9b136eb972da636726525bffc1c2f86fcb432))

## [0.3.1](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.3.0...modules/core/v0.3.1) (2026-01-09)


### 🐛 Bug Fixes

* Wrong rule checked in scheduled rollout ([#4471](https://github.com/thomaspoignant/go-feature-flag/issues/4471)) ([8eaf4ac](https://github.com/thomaspoignant/go-feature-flag/commit/8eaf4acf1e427cfc828457500b0ab44d3f634a8b))


### 🔧 Chores

* Bump k8s.io/api from 0.34.2 to 0.34.3 ([#4464](https://github.com/thomaspoignant/go-feature-flag/issues/4464)) ([7c947e6](https://github.com/thomaspoignant/go-feature-flag/commit/7c947e68ec359dbf2afbb1ac80ccbd31c56982cb))
* Bump k8s.io/api from 0.34.3 to 0.35.0 ([#4517](https://github.com/thomaspoignant/go-feature-flag/issues/4517)) ([631144b](https://github.com/thomaspoignant/go-feature-flag/commit/631144b01a33c2531ab4d2160be908d96a80347f))
* Code cleaning following sonar recommendation - part 2 ([#4470](https://github.com/thomaspoignant/go-feature-flag/issues/4470)) ([cee1dd7](https://github.com/thomaspoignant/go-feature-flag/commit/cee1dd71a571da3b8048b1469f21c7a251466766))
* fix some tests ([#4475](https://github.com/thomaspoignant/go-feature-flag/issues/4475)) ([2346f6e](https://github.com/thomaspoignant/go-feature-flag/commit/2346f6e179db302d61ca5fb6aa6ab26f577970ca))
* **test:** improve test coverage for core modules ([#4476](https://github.com/thomaspoignant/go-feature-flag/issues/4476)) ([7d96999](https://github.com/thomaspoignant/go-feature-flag/commit/7d96999d7a947d45e2e42df3d7fa88169fbbe9c9))

## [0.3.0](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.2.0...modules/core/v0.3.0) (2025-12-09)


### 🚀 New Features

* allow empty evaluation context for flags that don't require bucketing ([#3962](https://github.com/thomaspoignant/go-feature-flag/issues/3962)) ([0564b06](https://github.com/thomaspoignant/go-feature-flag/commit/0564b0680ec6da62bd012fbb3cafa8fb20d20d2c))
* Support x-api-key header for authentication ([#4347](https://github.com/thomaspoignant/go-feature-flag/issues/4347)) ([3ca07a8](https://github.com/thomaspoignant/go-feature-flag/commit/3ca07a8fa49522aa8b348bb5314a6f503dfa9778))


### 🔧 Chores

* Bump github.com/aws/aws-sdk-go-v2/service/s3 ([#4263](https://github.com/thomaspoignant/go-feature-flag/issues/4263)) ([3944a49](https://github.com/thomaspoignant/go-feature-flag/commit/3944a491413056d903236573fbf5a75fc7336dd9))
* Bump github.com/aws/aws-sdk-go-v2/service/s3 ([#4326](https://github.com/thomaspoignant/go-feature-flag/issues/4326)) ([927e392](https://github.com/thomaspoignant/go-feature-flag/commit/927e392662eaad75f33bd88275c566a465d02446))

## [0.2.0](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.1.4...modules/core/v0.2.0) (2025-11-03)


### 🚀 New Features

* Support nested property access in bucketingKey ([#4198](https://github.com/thomaspoignant/go-feature-flag/issues/4198)) ([284638d](https://github.com/thomaspoignant/go-feature-flag/commit/284638d019eee39a93aee213dcc729ce9ebcd33f))

## [0.1.4](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.1.3...modules/core/v0.1.4) (2025-10-14)


### 🔧 Chores

* adding missing comments ([#4090](https://github.com/thomaspoignant/go-feature-flag/issues/4090)) ([2ca2369](https://github.com/thomaspoignant/go-feature-flag/commit/2ca2369d16ede4a5bcf0206fd71e1fb8eed7fd0c))

## [0.1.3](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.1.2...modules/core/v0.1.3) (2025-10-14)


### 🔧 Chores

* adding missing comments ([#4087](https://github.com/thomaspoignant/go-feature-flag/issues/4087)) ([6a3ba2d](https://github.com/thomaspoignant/go-feature-flag/commit/6a3ba2df51d3ee0248943b79042132050b8ca876))

## [0.1.2](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.1.1...modules/core/v0.1.2) (2025-10-14)


### 🔧 Chores

* adding comments ([#4084](https://github.com/thomaspoignant/go-feature-flag/issues/4084)) ([1be1312](https://github.com/thomaspoignant/go-feature-flag/commit/1be131211ad85fc072ad5125fe8e5c87590711b9))

## [0.1.1](https://github.com/thomaspoignant/go-feature-flag/compare/modules/core/v0.1.0...modules/core/v0.1.1) (2025-10-14)


### 🔧 Chores

* **repo:** Create an evaluation go module ([#4079](https://github.com/thomaspoignant/go-feature-flag/issues/4079)) ([2305959](https://github.com/thomaspoignant/go-feature-flag/commit/230595939b35e9472a422e0b265fb450b20d3651))

## 0.1.0 (2025-10-10)


### 🔧 Chores

* initial release please setup ([#4028](https://github.com/thomaspoignant/go-feature-flag/issues/4028)) ([081e1ab](https://github.com/thomaspoignant/go-feature-flag/commit/081e1aba45f7d32073802ddceb3790766c6ef4ea))
