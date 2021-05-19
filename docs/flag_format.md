# Configure a flag
The goal of this module is to avoid having to host a backend to manage your feature flags and, to keep them centralized by using a single file as a source.  

Your file must be a valid `YAML`, `JSON` or `TOML` file with a list of flags *(examples: [`YAML`](https://github.com/thomaspoignant/go-feature-flag/tree/main/testdata/flag-config.yaml), [`JSON`](https://github.com/thomaspoignant/go-feature-flag/tree/main/testdata/flag-config.json), [`TOML`](https://github.com/thomaspoignant/go-feature-flag/tree/main/testdata/flag-config.toml))*.

## Example
A flag configuration looks like:

=== "YAML"

    ``` yaml linenums="1"
    test-flag:
      percentage: 100
      rule: key eq "random-key"
      true: true
      false: false
      default: false
      disable: false
      trackEvents: true
      rollout:
        experimentation:
          start: 2021-03-20T00:00:00.10-05:00
          end: 2021-03-21T00:00:00.10-05:00

    test-flag2:
      rule: key eq "not-a-key"
      percentage: 100
      true: true
      false: false
      default: false
    ```

=== "JSON"

    ``` json linenums="1"
    {
      "test-flag": {
        "percentage": 100,
        "rule": "key eq \"random-key\"",
        "true": true,
        "false": false,
        "default": false,
        "disable": false,
        "trackEvents": true,
        "rollout": {
          "experimentation": {
            "start": "2021-03-20T05:00:00.100Z",
            "end": "2021-03-21T05:00:00.100Z"
          }
        }
      },
      "test-flag2": {
        "rule": "key eq \"not-a-key\"",
        "percentage": 100,
        "true": true,
        "false": false,
        "default": false
      }
    }
    ```

=== "TOML"

    ``` toml linenums="1"
    [test-flag]
    percentage = 100.0
    rule = "key eq \"random-key\""
    true = true
    false = false
    default = false
    disable = false
    trackEvents = true
    
    [test-flag.rollout]

        [test-flag.rollout.experimentation]
        start = 2021-03-20T05:00:00.100Z
        end = 2021-03-21T05:00:00.100Z

    [test-flag2]
    rule = "key eq \"not-a-key\""
    percentage = 100.0
    true = true
    false = false
    default = false
    ```

## Format details
| Field | Description |
|:---:|---|
| **flag-key** | The `flag-key` is the name of your flag.<br> It must be unique.<br>*On the example the flag keys are **`test-flag`** and **`test-flag2`**.*|
| `true` | The value return by the flag if apply to the user *(rule is evaluated to true)* and user is in the active percentage.|
| `false`| The value return by the flag if apply to the user *(rule is evaluated to true)* and user is **not** in the active percentage.|
| `default` |The value return by the flag if not apply to the user *(rule is evaluated to false).*|
| `percentage` |*(optional)*<br>Percentage of the users affect by the flag.<br>**Default: 0**<br><br>The percentage is compute by doing a hash of the user key *(100000 variations)*, it means that you can have 3 numbers after the comma.|
| `rule` |*(optional)*<br>This is the query use to select on which user the flag should apply.<br>Rule format is describe in the <a href="#rule-format">rule format section</a>.<br>**If no rule set, the flag apply to all users *(percentage still apply)*.**|
| `disable` |*(optional)*<br>True if the flag is disabled.<br>**Default: `false`**|
| `trackEvents` |*(optional)*<br>False if you don't want to export the data in your data exporter.<br>**Default: `true`**|
| `rollout` |*(optional)*<br><code>rollout</code> contains a specific rollout strategy you want to use.<br>**See [rollout section](rollout/index.md) for more details.**|


## Rule format
The rule format is based on the [`nikunjy/rules`](https://github.com/nikunjy/rules) library.

All the operations can be written capitalized or lowercase (ex: `eq` or `EQ` can be used).  
Logical Operations supported are `AND` `OR`.

Compare Expression and their definitions (`a|b` means you can use either one of the two `a` or `b`):


| Operator | Description |
|:---:|---|
|`eq` \| `==`| equals to|
|`ne` \| `!=`| not equals to|
|`lt` \| `<` | less than|
|`gt` \| `>` | greater than|
|`le` \| `<=` | less than equal to|
|`ge` \| `>=` | greater than equal to| 
|`co` | contains | 
|`sw` | starts with| 
|`ew` | ends with|
|`in` | in a list|
|`pr` | present|
|`not` | not of a logical expression |

### Examples

- Select a specific user: `key eq "example@example.com"`
- Select all identified users: `anonymous ne true`
- Select a user with a custom property: `userId eq "12345"`

## Advanced configurations

You can have advanced configurations for your flag to have specific behavior for them, such as:

- [Specific rollout strategies](rollout/index.md)
- [Don't track a flag](data_collection/index.md#dont-track-a-flag)
