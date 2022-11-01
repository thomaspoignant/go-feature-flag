---
sidebar_position: 60
description: Description of the available endpoints in the relay proxy.
---

# Relay proxy endpoints

The most updated documentation about the relay proxy endpoints is the Swagger docs _(see [Swagger section](#swagger) to see how to access to the documentation)_.

## Specific to Relay Proxy	
	
### Health (health check)
Making a `GET` request to the URL path `/health` will tell you if the **relay proxy** is ready to serve traffic.  
This is useful especially for loadbalancer to know that they can send traffic to the service.

```json
{ "initialized": true }
```

### Info
Making a `GET` request to the URL path `/info` will tell give you information about the actual state of the **relay proxy**.  
As of Today the level of information is small be we can improve this endpoint to returns more information.

```json
{
    "cacheRefresh": "2022-06-13T11:22:55.941628+02:00"
}
```

| Field name         | Type | Description                                                                         |
|--------------------|------|-------------------------------------------------------------------------------------|
| `cacheRefresh`     | date | This is the last time when your flag file was read and store in the internal cache. |


### Swagger
Swagger endpoint is serving a [swagger UI](https://swagger.io/tools/swagger-ui/) to test your REST endpoints.
By default, this endpoint is not exposed, you need to have this configuration in your **relay proxy** configuration file:

```yaml
# ...
enableSwagger: true
host: my-proxy-domain.com # the DNS to access the proxy
```

When enabled, you can go to the `/swagger/` endpoint with your browser, and you will have access to the Swagger UI for the relay proxy. 

## Proxies for GO Feature Flag services

### Endpoint to get variation for a flag

Making a `POST` request to the URL `/v1/feature/<your_flag_name>/eval` will give you the value of the flag for this user.  
To get a variation you should provide information about the user.  
For that you should provide some user information in `JSON` in the request body.

#### Request
**Example:**
```json
{
  "user": {
    "key": "123e4567-e89b-12d3-a456-426614174000",
    "anonymous": false,
    "custom": {
      "firstname": "John",
      "lastname": "Doe",
      "email": "john.doe@random.io"
    }
  },
  "defaultValue": "default_value_provided_by_SDK"
}
```

| Field name         | Type               | Description                                                                                |
|--------------------|--------------------|--------------------------------------------------------------------------------------------|
| `user`             | [user](#user_type) | **(mandatory)** The representation of a user for your feature flag system.                 |
| `defaultValue`     | any                | **(mandatory)** The value will we use if we are not able to get the variation of the flag. |


#### Response

Based on the name of the flag and the user you will always have a response for the variation.  
The API will respond with a `200` even if the flag is not available, the goal is for your application to not crash even if
the flag does not exist anymore.

**Example:**
```json
{
  "value": "welcome_new_feature",
  "variationType": "true",
  "version": "0",
  "trackEvents": true,
  "failed": false
}
```

<a name="variation_results_details"></a>

| Field name      | Type    | Description                                                                                                        |
|-----------------|---------|--------------------------------------------------------------------------------------------------------------------|
| `value`         | any     | The flag value for this user.                                                                                      |
| `variationType` | string  | The variation used to give you this value.                                                                         |
| `version`       | string  | The version of the flag used.                                                                                      |
| `trackEvents`   | boolean | `true` if the event was tracked by the relay proxy.                                                                |
| `failed`        | boolean | `true` if something went wrong in the relay proxy _(flag does not exists, ...)_ and we serve the **defaultValue**. |



### Endpoint to get all flags variations for a user

Making a `POST` request to the URL `/v1/allflags` will give you the values of all the flags for this user.  
To get a variation you should provide information about the user.  
For that you should provide some user information in `JSON` in the request body.

#### Request
**Example:**
```json
{
  "user": {
    "key": "123e4567-e89b-12d3-a456-426614174000",
    "anonymous": false,
    "custom": {
      "firstname": "John",
      "lastname": "Doe",
      "email": "john.doe@random.io"
    }
  }
}
```

| Field name         | Type               | Description                                                                                |
|--------------------|--------------------|--------------------------------------------------------------------------------------------|
| `user`             | [user](#user_type) | **(mandatory)** The representation of a user for your feature flag system.                 |

#### Response

With the input user the API will loop over all flags and get values for all of them.

**Example:**
```json
{
  "flags": {
    "flag-only-for-admin": {
      "value": false,
      "timestamp": 1655123971,
      "variationType": "Default",
      "trackEvents": true
    },
    "new-admin-access": {
      "value": false,
      "timestamp": 1655123971,
      "variationType": "False",
      "trackEvents": true
    }
  },
  "valid": true
}
```

| Field name    | Type                       | Description                                                                                                                                          |
|---------------|----------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| `flags`       | map[string]variationResult | All the flags with their results _(see [Endpoint to get all flags variations for a user](#variation_results_details) for the details of the format)_ |
| `valid`       | boolean                    | `true` if something went wrong in the relay proxy _(flag does not exists, ...)_ and we serve the **defaultValue**.                                   |

<a name="user_type"></a>

## User type

This type represent in your request body the information about a user.  
This is based on these information that we will be able to filter which variation apply for this user. 

```json
{
  "key": "123e4567-e89b-12d3-a456-426614174000",
  "anonymous": false,
  "custom": {
    "firstname": "John",
    "lastname": "Doe",
    "email": "john.doe@random.io"
  }
}
```

| Field name  | Type                   | Default   | Description                                                                                                                            |
|-------------|------------------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------|
| `key`       | string                 | **none**  | **(mandatory)** Unique key of your user, it could be any string, I recommend to use UUID, email or whatever who make your user unique. |
| `anonymous` | boolean                | `false`   | Is it an authenticated user or not.                                                                                                    |
| `custom`    | map[string]interface{} | **empty** | This is an object where you can put everything you think is useful, you will be able to use rule based on these fields.                |

