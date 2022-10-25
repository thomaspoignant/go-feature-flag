# Canary Release

**Canary release** is a technique to reduce the risk of introducing a new software version in production by slowly rolling out the change to a small subset of users before rolling it out to the entire infrastructure and making it available to everybody.

This is the easiest rollout strategy available.  
You just have to select a percentage of your users in your flag, and the `True` behavior will apply to them.

## Example

### YAML

``` yaml linenums="1"
canary-flag:
  true: true
  false: false
  default: false
  # highlight-next-line
  percentage: 1
```

### JSON

``` json
{
  "canary-flag": {
    "true": true,
    "false": false,
    "default": "false,
    # highlight-next-line
    "percentage": 1
  }
}
```

### TOML

``` toml
[canary-flag]
true = true
false = false
default = false
# highlight-next-line
percentage = 1.0
```
