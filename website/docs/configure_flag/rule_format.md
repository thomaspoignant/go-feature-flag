---
sidebar_position: 21
description: How to create a rule to target specific users
---

# How to target specific users

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
      <td>Name of your rule.<br/>This is needed when your are updating a rule using a <a href="./rollout/scheduled">scheduled rollout</a>.</td>
    </tr>
    <tr>
      <td><code>query</code></td>
      <td>
        <p>
          Query represents an <b>antlr query</b> in the <b>nikunjy/rules</b> format.
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
          <pre>
            percentage:<br/>  variationA: 10.59<br/>  variationB: 9.41<br/>  variationC: 80
          </pre>
        <p>The format is the name of the variation and the percentage for this one.</p>
        <p><b>Note: If your total is not equal to 100%, this rule will be considered invalid.</b></p>
      </td>
    </tr>
    <tr>
      <td><code>progressiveRollout</code><br/><i>(optional)</i></td>
      <td>
        <p>
          Allows you to ramp up the percentage of your flag over time.
        </p>
        <p>
          You can decide at which percentage you start and end with in your release ramp.
          Before the start date we will serve the initial percentage and afterwards, we will serve the end percentage.
        </p>
        <p><i>See <a href="./rollout/progressive">progressive rollout</a> to have more info on how to use it.</i></p>
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


:::info
`variation`, `percentage` and `progressiveRollout` are optional but you **must have at least one of the three**.

If you have more than one field we will use the first one in the order
`progressiveRollout` > `percentage` > `variation`.
:::

### Query format

The rule format is based on the [`nikunjy/rules`](https://github.com/nikunjy/rules) library.

All the operations can be written in capitalized or lowercase characters (ex: `eq` or `EQ` can be used).
Logical Operations supported are `AND` & `OR`.

Compare Expression and their definitions (`a|b` means you can use one of either `a` or `b`):

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
  (key ew "@test.com") and (role eq "backend engineer") and (env eq "pro") and (company eq "go-feature-flag")
  ```

## Environments

When you initialise `go-feature-flag` you can set an [environment](../go_module/configuration/#option_environment) for the instance of this SDK.

```go linenums="1"
ffclient.Init(ffclient.Config{
    // ...
    Environment:    "prod",
    // ...
})
```

When an environment is set, it adds a new field in your user called **`env`** that you can use in your queries.
It means that you can decide to activate a flag only for some **environment**.

**Example of flag configuration based on the environment:**

```yaml
my-flag:
  variations:
    A: "A"
    B: "B"
    C: "C"
  targeting:
    - name: Target pre environment
      query: env eq "pre"
      variation: A
    - name: Target pro environment
      query: env eq "pro"
      variation: B
  defaultRule:
    variation: C
```

## Get the rule name in the metadata

When you use a rule in your targeting, you can get the name of the rule in the metadata of the variation.  
The information on what rule has been used to serve the variation is available in the metadata of the variation in the field called `evaluatedRuleName`.

If you are interested about this information, you have to name your rules by adding the field `name` in your rule. This name will be extract and added in the `evaluatedRuleName` field of the metadata.
