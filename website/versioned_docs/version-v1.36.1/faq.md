---
sidebar_position: 100
---
# Frequently Asked Questions

### Why Use Feature Flags?
This one of most common question I get.
Feature flags are a software development technique that allows the toggling of specific functionalities on and off at runtime without the need to deploy new code.

It allows you to decouple **deploy** and **release**, giving you better control and more experimentation over the full lifecycle of features.

---

### What is the lifecycle of a flag?
Managing the lifecycle of feature flags is crucial to prevent cluttering your codebase with obsolete elements. Here's a step-by-step guide:

1. **Creation**: Initiate by adding the flag to your configuration file, setting it to 0% to avoid impacting users.
2. **Evaluation**: Implement the flag evaluation in your code (refer to [variation](./go_module/target_user.md#variation)).
3. **Deployment**: Deploy your application with the variation check in place.
4. **Rollout**: Gradually enable the flag for users.
5. **Completion**: Once the feature reaches 100% visibility, eliminate the variation call from your code.
6. **Clean-Up**: Deploy your application sans the variation check.
7. **Removal**: Finally, delete the flag from your configuration file.

---

### What happens if my configuration file is not reachable/deleted?
If while you are on production and for some reason your flag file becomes unreachable, we will be able to serve the users based on the last version of the file we were able to read. We will continue to try reading the file based on the `pollingInterval` you have configured.

If you start a new instance and the file is not reachable to module, it will fail to initialize except if you have set the option `StartWithRetrieverError` in the config. With this option, we will serve the SDK the default value *(the 3rd param in your variation)* until the flag becomes available again.

---

### What is the best rollout strategy?
The lib offers numerous rollout strategies, with no single "best" approach as it heavily depends on the context of your feature release. 
Some strategies include:

- **Simple Cut-Off**: For non-critical releases, transitioning the flag from 0% to 100% immediately for all users might be suitable.
- **Progressive Rollout**: For releases that might impact infrastructure, a gradual rollout can mitigate risks by incrementally increasing user exposure.
- **Targeted Release**: To affect only a specific user segment, applying rules to your flag can be effective.

You have an endless list of rollout strategies depending on what is your feature.

---

### How do we ensure that users affected by the feature flags are not always the same?

To avoid always have the same users getting affected by a flag, we compute the hash that allows us to determine if the user is part of the percentage that is not computed only based on the user key but a combination of the user key and the flag name.

It guarantees that the user will be always in the same group but depending on the flag.

---
