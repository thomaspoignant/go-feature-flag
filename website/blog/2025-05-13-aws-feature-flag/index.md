---
title: AWS Feature Flags, Everything You Should Know to Use GO Feature Flag on AWS
description: AWS Feature Flags, Everything You Should Know to Use GO Feature Flag on AWS and leverage all the integrations points.
authors: [thomaspoignant]
tags: [AWS, Amazon Web Services]
---

In the rapidly evolving landscape of cloud-native applications, feature flags have become an indispensable tool for modern development teams. They empower you to decouple code deployments from feature **releases**, **enabling safer rollouts**, **A/B testing**, **targeted feature delivery**, and **quick rollbacks** without redeploying your entire application.

For many organizations, the cloud platform of choice is Amazon Web Services (AWS). AWS offers a vast array of services that provide the scalability, reliability, and flexibility required for demanding applications. So, how do you bring the power of feature flags into your AWS-hosted applications?

Enter [GO Feature Flag](https://gofeatureflag.org) – an open-source, simple, and performant feature flag solution designed to integrate seamlessly into your existing infrastructure. In this post, we'll explore how you can effectively leverage GO Feature Flag with the AWS ecosystem to manage your features with precision and confidence.

## The AWS Ecosystem for Feature Flags: Your Building Blocks
AWS provides a feature flag solution with Apps Config, but this is really basic and will not give you the best experience with your flags.  
But the good thing is that you can leverage GO Feature Flag with existing AWS services to have a minimal setup and enjoy using feature flags on AWS.

AWS provides several services that are ideal candidates for storing and managing your feature flag configurations. Understanding these options is crucial for designing a robust and scalable feature flagging system.

## Data Storage & Configuration: Where Your Flags Live
GO Feature Flag is not necesseraly using a database to store your feature flags and can work with a simple file stored in your preffered solution.

If you want to be fully integrated with AWS, AWS S3 is probably your best choice when it comes to storing your flags. S3 is a highly scalable, durable, and cost-effective object storage service. It's an excellent choice for storing your feature flag configurations as YAML files.
- Benefits: Extremely high availability, low cost, easy to integrate for static or infrequently changing flags. You can also version your flag files on S3.




    HTTP(S)
    File System
    Kubernetes ConfigMap
    AWS S3
    Google Cloud Storage
    Azure Blob Storage
    GitHub
    GitLab
    Bitbucket
    MongoDB
    Redis