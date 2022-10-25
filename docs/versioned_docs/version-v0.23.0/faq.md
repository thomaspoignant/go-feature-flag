# Frequently Asked Questions

### Why using feature flags?
This one of most common question I get.
Feature flags are a software development technique that turns certain functionality on and off during runtime, without 
deploying new code.

It allows you to decouple **deploy** and **release**, giving you better control and more experimentation over the full 
lifecycle of features.

---

### What is the lifecycle of a flag?
The lifecycle of your flags is key if you don't want to have un-used things everywhere in your code. 

1. Start by creating the flag in your configuration file *(with 0% to avoid affecting your users)*.
2. Evaluate the flag in your code *(see [variation](users.md#variation))*.
3. Deploy your application with the variation check.
4. Start rolling out your flag.
5. When 100% of your users have access to the new feature, remove the call to the variation from your code base.
6. Deploy your application without the variation check.
7. Remove the flag from your configuration file.

---

### What happen if my configuration file is not reachable/deleted?
If while you are on production for some reason your flag file becomes unreachable, we will be able to serve the users
based on the last version of the file we were able to read. We will continue to try reading the file based on
the `pollingInterval` you have configured.

If you start a new instance, if the file is not reachable to module will fail to initialize except if you have set the 
option `StartWithRetrieverError` in the config. With this option, we will serve the SDK default value *(the 3rd param
in your variation)* until the flag becomes available again.

---

### What is the best rollout strategy?
The lib gives you a lot of strategies to rollout your flags, there is no better one, it always depends on the context
of your release.

- If your release is not critical and, you just want an easy cut-off strategy, you can pass your flag from 0% to 100% for 
    all your users
- If you are scared that your infrastructure can be impacted by the new feature, a progressive rollout can help you to
    impact users over time and to be able to check how your system handle it.
- If you want to impact only a subset of your users, you can put a rule on your flag.
- Etc ...

You have an endless list of rollout strategies depending on what is your feature.

---

### How we ensure that users affected by the feature flags are not always the same?

To avoid always have the same users affected by a flag, the hash we compute that allows us to determine if the user is part of the percentage is not computed only based on the user key but a combination of the user key and the flag name.

It guarantees that the user will be always in the same group but depending on the flag.

---
