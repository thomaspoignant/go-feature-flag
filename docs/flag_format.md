---
sidebar_position: 4
---

# Configure a flag

`go-feature-flag` core feature is to centralize all your feature flags in a source file, and to avoid hosting and maintaining a backend server to manage them.  

Your file must be a valid `YAML`, `JSON` or `TOML` file with a list of flags
*(examples: [`YAML`](https://github.com/thomaspoignant/go-feature-flag/tree/main/testdata/flag-config.yaml),
[`JSON`](https://github.com/thomaspoignant/go-feature-flag/tree/main/testdata/flag-config.json),
[`TOML`](https://github.com/thomaspoignant/go-feature-flag/tree/main/testdata/flag-config.toml))*.

The easiest way to create your configuration file is to used
[**GO Feature Flag Editor** available at [https://editor.gofeatureflag.org](https://editor.gofeatureflag.org).  
If you prefer to do it manually please follow instruction bellow.

## Editor

Creating the first version of the flag is not always pleasant, that's why we have created
[**GO Feature Flag Editor**](https://editor.gofeatureflag.org) a simple editor to create your files.  

## Example

A flag configuration looks like:

=== "YAML"

``` yaml linenums="1"
# This is your configuration for your first flag
first-flag:
  variations: # All possible return value for your feature flag
    A: false
    B: true
  targeting: # If you want to target a subset of your users in particular
    - query: key eq "random-key"
      percentage:
        A: 0
        B: 100
  defaultRule: # When no targeting match we use the defaultRule
    variation: A

# A second example of a flag configuration
second-flag:
  variations:
    A: "valueA"
    B: "valueB"
    defaultValue: "a default value"
  targeting:
    - name: legacyRuleV0
      query: key eq "not-a-key"
      percentage:
        A: 10
        B: 90
  defaultRule:
    name: legacyDefaultRule
    variation: defaultValue
  version: "12"
  experimentation: 
    start: 2021-03-20T00:00:00.1-05:00
    end: 2021-03-21T00:00:00.1-05:00
```

=== "JSON"

``` json linenums="1"
{
  "first-flag": {
    "variations": {
      "A": false,
      "B": true
    },
    "targeting": [
      {
        "query": "key eq \"random-key\"",
        "percentage": {
          "A": 0,
          "B": 100
        }
      }
    ],
    "defaultRule": {
      "variation": "A"
    }
  },
  
  "second-flag": {
    "variations": {
      "A": "valueA",
      "B": "valueB",
      "defaultValue": "a default value"
    },
    "targeting": [
      {
        "name": "legacyRuleV0",
        "query": "key eq \"not-a-key\"",
        "percentage": {
          "A": 10,
          "B": 90
        }
      }
    ],
    "defaultRule": {
      "name": "legacyDefaultRule",
      "variation": "defaultValue"
    },
    "version": "12",
    "experimentation": {
      "start": "2021-03-20T05:00:00.100Z",
      "end": "2021-03-21T05:00:00.100Z"
    }
  }
}
```

=== "TOML"

``` toml linenums="1"
[first-flag.variations]
A = false
B = true

[[first-flag.targeting]]
query = 'key eq "random-key"'

  [first-flag.targeting.percentage]
  A = 0
  B = 100

[first-flag.defaultRule]
variation = "A"

[second-flag]
version = "12"

  [second-flag.variations]
  A = "valueA"
  B = "valueB"
  defaultValue = "a default value"

  [[second-flag.targeting]]
  name = "legacyRuleV0"
  query = 'key eq "not-a-key"'

    [second-flag.targeting.percentage]
    A = 10
    B = 90

  [second-flag.defaultRule]
  name = "legacyDefaultRule"
  variation = "defaultValue"

  [second-flag.experimentation]
  start = 2021-03-20T05:00:00.100Z
  end = 2021-03-21T05:00:00.100Z
```

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
      <td width="20%"><strong>flag-key</strong></td>
      <td>Name of your flag.<br/><i>It must be unique.<br/>On the example the flag keys are <code>test-flag</code> and <code>test-flag2</code>.</i></td>
    </tr>
    <tr>
      <td><code>variations</code></td>
      <td>
        <p>Variations are all the variations available for this flag.</p><p>It is represented as a key/value element. The key is the name of the variation and the value could be any types available (<code>string</code>, <code>float</code>, <code>int</code>, <code>map</code>, <code>array</code>, <code>bool</code>).</p><p>You can have as many variation as needed.</p>
      
         ```yaml
                 # Some examples
                 variationString: test
                 variationBool: true
                 variationInt: 1000
                 variationFloat: 1000.23
                 variationArray: 
                   - item1
                   - item2
                 variationObj:
                   item1: 123
                   item2: this is a string
                   item3: false
         ```
      
      </td>
    </tr>
    <tr>
      <td><code>targeting</code><br/><i>(optional)</i></td>
      <td>
        <p>Targeting contains the list of rules you have to target a subset of your users.<br/>You can have as many target as needed.</p>
        <p>This field is an <code>array</code> and contains a list of rules. </p>
        <p><i>See <a href="#rule-format">rules format</a> to have more info on how to write a rule.</i></p>
      </td>
    </tr>
    <tr>
      <td><code>defaultRule</code></td>
      <td>
        <p>DefaultRule is the rule that is applied if the user does not match in any targeting.</p>
        <p><i>See <a href="#rule-format">rules format</a> to have more info on how to write a rule.</i></p>
      </td>
    </tr>
    <tr>
      <td><code>trackEvents</code><br/><i>(optional)</i></td>
      <td>
        <p><code>false</code> if you don't want to export the data in your data exporter.</p>
        <p><b>Default:</b> <code>true</code></p>
      </td>
    </tr>
    <tr>
      <td><code>disable</code><br/><i>(optional)</i></td>
      <td>
        <p><code>true</code> if the flag is disabled.</p>
        <p><b>Default:</b> <code>false</code></p>
      </td>
    </tr>
    <tr>
      <td><code>version</code><br/><i>(optional)</i></td>
      <td>
        <p>The version is the version of your flag.<br/>This string is used to display the information in the notifiers and data collection, you have to update it your self.</p>
        <p><b>Default:</b> <code>""</code></p>
      </td>
    </tr>
    <tr>
      <td><code>scheduledRollout</code><br/><i>(optional)</i></td>
      <td>
        <p>Scheduled allow to patch your flag over time.</p>
        <p>You can add several steps that updates the flag, this is typically used if you want to gradually add more user in your flag.</p>
        <p><i>See <a href="../rollout/scheduled/">Scheduled rollout</a> to have more info on how to use it.</i></p>
      </td>
    </tr>
    <tr>
      <td><code>experimentation</code><br/><i>(optional)</i></td>
      <td>
        <p>Experimentation allow you to configure a start date and an end date for your flag. When the experimentation is not running, the flag will serve the default value.</p>
        <p><i>See <a href="../rollout/experimentation/">Experimentation rollout</a> to have more info on how to use it.</i></p>
      </td>
    </tr>
  </tbody>
</table>


## Rule format

A rule is a configuration that allows to serve a variation based on some conditions.

### Format details

<table>
  <thead>
    <tr>
      <th width="20%">Field</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>name</code><br/><i>(optional)</i></td>
      <td>Name of your rule.<br/>This is needed when your are updating a rule using a <a href="../rollout/scheduled">scheduled rollout</a>.</td>
    </tr>
    <tr>
      <td><code>query</code></td>
      <td>
        <p>
          Query represents an antlr query in the nikunjy/rules format.
          <br/><b>This field is mandatory in every rule used in the targeting field</b>.
        </p>
        <p><i>See <a href="#query-format">query format</a> to have the syntax.</i></p>
        <p><i>Note: if you use the field <code>query</code> in a <code>defaultRule</code> it will be ignored.</i></p>
      </td>
    </tr>
    <tr>
      <td><code>variation</code><br/><i>(optional)</i></td>
      <td>Name of the variation to return.</td>
    </tr>
    <tr>
      <td><code>percentage</code><br/><i>(optional)</i></td>
      <td>
        <p>Represents the percentage we should give to each variation.</p>
        <p>
        
        ```yaml
        percentage:
          variationA: 10.59
          variationB: 9.41
          variationC: 80
        ```        

        </p>
        <p>The format is the name of the variation and the percentage for this one.</p>
        <p><b>Note: if your total is not equals to 100% this rule will be considered as invalid.</b></p>     


      </td>
    </tr>
    <tr>
      <td><code>progressiveRollout</code><br/><i>(optional)</i></td>
      <td>
        <p>Allow to ramp up the percentage of your flag over time.</p>
        <p>You can decide at which percentage you starts with and at what percentage you ends with in your release ramp.
          Before the start date we will serve the initial percentage and, after we will serve the end percentage.
        </p>
        <p><i>See <a href="../rollout/progressive">progressive rollout</a> to have more info on how to use it.</i></p>


      </td>
    </tr>
    <tr>
      <td><code>disable</code><br/><i>(optional)</i></td>
      <td>
        <p>Set to <code>true</code> if you want to disable the rule.</p>
        <p><b>Default:</b> <code>true</code>.</p>
      </td>
    </tr>
  </tbody>
</table>


!!! Info
    `variation`, `percentage` and `progressiveRollout` are optional but you need to have one of the 3.  
    
    If you have more than one field we will use the first one in that order  
    `progressiveRollout` > `percentage` > `variation`.

### Query format

The rule format is based on the [`nikunjy/rules`](https://github.com/nikunjy/rules) library.

All the operations can be written capitalized or lowercase (ex: `eq` or `EQ` can be used).  
Logical Operations supported are `AND` `OR`.

Compare Expression and their definitions (`a|b` means you can use either one of the two `a` or `b`):

|  Operator  | Description                 |
|:----------:|-----------------------------|
| `eq`, `==` | equals to                   |
| `ne`, `!=` | not equals to               |
| `lt`, `<`  | less than                   |
| `gt`, `>`  | greater than                |
| `le`, `<=` | less than equal to          |
| `ge`, `>=` | greater than equal to       |
|    `co`    | contains                    |
|    `sw`    | starts with                 |
|    `ew`    | ends with                   |
|    `in`    | in a list                   |
|    `pr`    | present                     |
|   `not`    | not of a logical expression |

#### Examples

- Select a specific user: `key eq "example@example.com"`
- Select all identified users: `anonymous ne true`
- Select a user with a custom property: `userId eq "12345"`
- Select on multiple criteria:  
  *All users with ids finishing by `@test.com` that have the role `backend engineer` in the `pro` environment for the
  company `go-feature-flag`*

  ```bash
  (key ew "@test.com") and (role eq "backend engineer") and (env eq "pro") and (company eq "go-feature-flag")`
  ```

## Environments

When you initialise `go-feature-flag` you can set an [environment](../configuration/#option_environment) for the instance of this SDK.

```go linenums="1"
ffclient.Init(ffclient.Config{ 
    // ...
    Environment:    "prod",
    // ...
})
```

When an environment is set, it adds a new field in your user called **`env`** that you can use in your rules.  
It means that you can decide to activate a flag only for some **environment**.

**Example of rules based on the environment:**

```yaml
# Flag activate only in dev
rule: env == "dev"
```

```yaml
# Flag used only in dev and staging environment
rule: (env == "dev") or (env == "staging")
```

```yaml
# Flag used on non prod environments except for the user 1234 in prod
rule: (env != "prod") or (user_id == 1234)
```

## Advanced configurations

You can have advanced configurations for your flag to have specific behavior for them, such as:

- [Specific rollout strategies](rollout/index.md)
- [Don't track a flag](data_collection/index.md#dont-track-a-flag)
