# Playground API
The editor API is used to evaluate your flags for testing purposes.

It has been built to test your configuration in the [Flag editor](https://editor.gofeatureflag.org).

## Flag Test API
This API is testing an evaluation context with a flag configuration.  
All the information are expected in JSON.

### Example of request

```json
{
  "context": {
    "key": "aae1cb41-c3cb-4753-a117-031ddc958e81",
    "custom": {
      "name": "John Doe",
      "email": "john.doe@gofeatureflag.org"
    }
  },
  "flagName": "my-flag-name",
  "flag": {
    "variations": {
      "Default": false,
      "A": false,
      "B": true
    },
    "targeting": [
      {
        "query": "key eq \"aae1cb41-c3cb-4753-a117-031ddc958e81\"",
        "percentage": {
          "A": 0,
          "B": 100
        }
      }
    ],
    "defaultRule": {
      "variation": "Default"
    }
  }
}
```

### Example of response
```json
{
    "trackEvents": false,
    "variationType": "B",
    "failed": false,
    "version": "",
    "reason": "TARGETING_MATCH",
    "errorCode": "",
    "value": true,
    "cacheable": true
}
```