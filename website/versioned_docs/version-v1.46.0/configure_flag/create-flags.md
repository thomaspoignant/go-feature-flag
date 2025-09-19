---
sidebar_position: 10
description: What is a flag and how you can create them.
---

# 🆕 Create Flags

## Overview

In **GO Feature Flag** a flag is a configuration that allows you to serve a variation based on some conditions.

Since day 1, we have decided that we won't be the one telling you were to store your feature flags, and this is why we have introduce the concept of [**retriever**](../concepts/retriever.mdx).
To support this concept we had to define a simple way to define your feature flags.

In **GO Feature Flag** you can store your flags in a file, and we support `YAML`, `JSON` and `TOML` format.

:::note **OpenFeature Definition of a Feature Flag**
A mechanism that allows an Application Author to define alternative codepaths within a deployed piece of software, which is conditionally executed at runtime, based on a rule set.
:::

## Editor

Creating the first version of the flag is not always pleasant, that's why we have created
[**GO Feature Flag Editor**](https://editor.gofeatureflag.org) a simple editor to create your files.

## ✏️ Create your first feature flag manually

1. When starting to use GO Feature Flag, the first step will be to create the flag that will store your feature flags.
   **This file should be a `YAML`, `JSON` or `TOML` file.**

   ✏️ In this file you can start creating your feature flags.

2. First, **find a name** for your feature flag (this will be the key of your flag and **it must be unique**).
3. Define the **variations** that your flag can return.
   ```yaml title="flag-config.goff.yaml"
   display-banner:
     variations:
       enabled: true
       disabled: false
       # ...
   ```
4. Define a **default rule**, that will be serve when no targeting match.
   ```yaml title="flag-config.goff.yaml"
   display-banner:
     variations:
       enabled: true
       disabled: false
     defaultRule:
       variation: disabled
   ```
5. **🎉 Congrats you have your first feature flag created.**
   _This flag will return the variation `disabled` and the value `false` for all the users, but you can start targeting a specific group of users to return the variant `enabled`._
6. Now you can [store your flag](../integrations/store-flags-configuration) file in your favorite storage and start using it in your application.

### 👌 Allowed types

GO Feature Flags allow you to use custom variations to dynamically configure flag behavior.
Traditionally, feature flags return a boolean value: `true` or `false`. This works well for basic end user management and a kill switch.

In addition to **boolean** values, GO Feature Flag allows you to return **strings**, **integers**, **floats**, **JSON arrays**, and **JSON objects**.
This makes it possible to do configuration management via feature flags and manage plans/complex functionality.

:::tip **Never update the type of your feature flag**
We recommend to never change the type of your feature flag _(if you have a boolean flag, keep it as a boolean)_.
Why? Because apps trying to evaluate them in the different language will expect the type to stay constant.
:::

### 🔢 Multi-variant flags

Since GO Feature Flag is managing more than `boolean` we can have as many alternative variations as needed.
This is useful for A/B testing, permissions management, and other use cases where targeting a consistent group of users is required.

```yaml title="flag-config.goff.yaml"
scream-level-feature:
  variations:
    low: 'whisper'
    medium: 'talk'
    high: 'scream'
  defaultRule:
    variation: low
```

As you can see in the example above, we have a flag that can return 3 different values: `whisper`, `talk`, and `scream`.

:::danger
It is important that all variations have the same type, if not the flag will be considered invalid.
:::

### 💯 Percentages affectation

A reason to use feature flags is to dynamically affect the traffic to different variations.
We want a way to control the percentage of users that will see a specific variation.

If we take our example of a multi-variant flag, we can split the traffic in 3 groups and affect percentages for each variation in order to control who is seeing what.

```yaml title="flag-config.goff.yaml"
scream-level-feature:
  variations:
    low: 'whisper'
    medium: 'talk'
    high: 'scream'
  defaultRule:
    percentage:
      low: 10
      medium: 50
      high: 40
```

:::info **How does it work?**
It builds a hash with the flag name and the targeting key provided in the evaluation context in order to split the traffic in a consistent way.

It means that for a specific `targetingKey` the user will always see the same flag variation overtime.
:::

## Format details

<table>
  <thead>
    <tr>
      <th>Field</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td width="20%">
        <strong>flag-key</strong>
      </td>
      <td>
        Name of your flag.
        <br />
        <i>It must be unique.</i>
      </td>
    </tr>
    <tr>
      <td>
        <code>bucketing-key</code>
        <br />
        <i><sup><sup>optional</sup></sup></i>
      </td>
      <td>
        <p>
          Selects a key from the evaluation context that will be used in place
          of the <code>targetingKey</code> for the purpose of evaluating which
          variant to serve.

          If <code>bucketingKey</code> is set and the value is missing from the context, the flag will not be evaluated.
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>variations</code>
      </td>
      <td>
        <p>Variations are all the variations available for this flag.</p>
        <p>
          It is represented as a key/value element. The key is the name of the
          variation and the value could be of any types available (
          <code>string</code>, <code>float</code>, <code>int</code>,{" "}
          <code>map</code>, <code>array</code>, <code>bool</code>).
        </p>
        <p>You can have as many variations as needed.</p>
        <pre>
          <h2>Some examples:</h2>
          variationString: test
          variationBool: true
          variationInt: 1000
          variationFloat: 1000.23
          variationArray: <br /> - item1<br/> - item2
          variationObj:<br /> item1: 123<br /> item2: this is a string<br /> item3: false
        </pre>
      </td>
    </tr>
    <tr>
      <td>
        <code>targeting</code>
        <br />
        <sup><sup>optional</sup></sup>
      </td>
      <td>
        <p>
          Targeting contains the list of rules you have to target a subset of
          your users.
          <br />
          You can have as many target as needed.
        </p>
        <p>
          This field is an <code>array</code> and contains a list of rules.
        </p>
        <p>
          <i>
            See <a href="./target-with-flag">Target with Flag</a> to have more info on
            how to write a rule.
          </i>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>defaultRule</code>
      </td>
      <td>
        <p>
          DefaultRule is the rule that is applied if the user does not match in
          any targeting.
        </p>
        <p>
          <i>
            See <a href="./target-with-flag">Target with Flag</a> to have more info on
            how to write a rule.
          </i>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>trackEvents</code>
        <br />
        <sup><sup>optional</sup></sup>
      </td>
      <td>
        <p>
          <code>false</code> if you don't want to export the data in your data exporter.
        </p>
        <p>
          <b>Default:</b> <code>true</code>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>disable</code>
        <br />
        <sup><sup>optional</sup></sup>
      </td>
      <td>
        <p>
          <code>true</code> if the flag is disabled.
        </p>
        <p>
          <b>Default:</b> <code>false</code>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>version</code>
        <br />
        <sup><sup>optional</sup></sup>
      </td>
      <td>
        <p>
          The `version` is the version of your flag.
          <br />
          This string is used to display the information in the notifiers and
          data collection, you have to update it yourself.
        </p>
        <p>
          <b>Default:</b> <code>""</code>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>metadata</code>
        <br />
        <sup><sup>optional</sup></sup>
      </td>
      <td>
        <p>
          This field allows adding a wealth of information about a particular
          feature flag, such as a configuration URL or the originating Jira
          issue.
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>scheduledRollout</code>
        <br />
        <sup><sup>optional</sup></sup>
      </td>
      <td>
        <p>`scheduledRollout` allows to patch your flag over time.</p>
        <p>
          You can add several steps that update the flag, this is typically
          used if you want to gradually add more user in your flag.
        </p>
        <p>
          <i>
            See <a href="./rollout-strategies/scheduled/">Scheduled rollout</a> to have
            more info on how to use it.
          </i>
        </p>
      </td>
    </tr>
    <tr>
      <td>
        <code>experimentation</code>
        <br />
        <sup><sup>optional</sup></sup>
      </td>
      <td>
        <p>
          Experimentation allows you to configure a start date and an end date
          for your flag. When the experimentation is not running, the flag will
          serve the default value.
        </p>
        <p>
          <i>
            See <a href="./rollout-strategies/experimentation/">Experimentation rollout</a>{" "}
            to have more info on how to use it.
          </i>
        </p>
      </td>
    </tr>

  </tbody>
</table>

## Advanced configurations

You can have advanced configurations for your flag for them to have specific behavior, such as:

- [Target with Flags](./target-with-flags)
- [Specific rollout strategies](./rollout-strategies/)
- [Have a specific bucketing key](./custom-bucketing)

---

## Example of a flag lifecycle

In this example, we want to test the right screaming level we should have for the Monster.Inc company to be successful.
Considering that Today the level used is "whisper", this how we want our flag to act.

#### 1. Creating the flag

We are creating a new flag in our file called `flag-config.goff.yaml` and we are naming it `scream-level-feature`.

```yaml title="flag-config.goff.yaml"
scream-level-feature:
  variations:
    low: 'whisper'
    medium: 'talk'
    high: 'scream'
  defaultRule:
    # highlight-next-line
    variation: low
```

#### 2. Code deployment

After the flag is configured in GO Feature Flag, we can deploy our code that is evaluating the flag.

Since we are returning the same level as before, nothing is changing, we are happy about that 😁.

#### 3. Start testing the flag

Now we can start testing the flag, we can start by asking our product manager to test the new feature in production.

To achieve this we are targeting the product manager with his unique ID and we are returning the variation `high`.

```yaml title="flag-config.goff.yaml"
scream-level-feature:
  variations:
    low: 'whisper'
    medium: 'talk'
    high: 'scream'
  targeting:
    # highlight-start
    - query: targetingKey eq "aae1cb41-c3cb-4753-a117-031ddc958e81"
      variation: high
    # highlight-end
  defaultRule:
    variation: low
```

_We can iterate on this phase until the result is satisfying._

#### 4. Testing variations

Now that we know that the flag is working well, we can start testing the 3 different variations, to see which one perform best.

```yaml title="flag-config.goff.yaml"
scream-level-feature:
  variations:
    low: 'whisper'
    medium: 'talk'
    high: 'scream'
  defaultRule:
    # highlight-start
    percentage:
      low: 34
      medium: 33
      high: 33
    # highlight-end
```

#### 5. Set the new default value

After testing the different variations, we can decide to change the default value to the one that performed the best \_(here is it `high`).

```yaml title="flag-config.goff.yaml"
scream-level-feature:
  variations:
    low: 'whisper'
    medium: 'talk'
    high: 'scream'
  defaultRule:
    # highlight-start
    variation: high
    # highlight-end
```

#### 5. Remove the flag

After some time and the feature is now part of the product, we can remove the flag.
