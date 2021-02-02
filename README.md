<p align="center">
  <img width="250" height="238" src="logo.png" alt="go-feature-flag logo" />
</p>

# 🎛️ go-feature-flag [![Tweet](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/intent/tweet?text=I%27ve%20discovered%20go-feature-flag%20a%20great%20solution%20to%20easily%20managed%20feature%20flag%20in%20golang&url=https%3A%2F%2Fgithub.com%2Fthomaspoignant%2Fgo-feature-flag&via=thomaspoignant&hashtags=golang,featureflags,featuretoggle,go)


<p align="center">
    <a href="https://app.circleci.com/pipelines/github/thomaspoignant/go-feature-flag"><img src="https://img.shields.io/circleci/build/github/thomaspoignant/go-feature-flag" alt="Build Status" /></a>
    <a href="https://coveralls.io/github/thomaspoignant/go-feature-flag"><img src="https://coveralls.io/repos/github/thomaspoignant/go-feature-flag/badge.svg" alt="Coverage Status" /></a>
    <a href="https://sonarcloud.io/dashboard?id=thomaspoignant_go-feature-flag"><img src="https://sonarcloud.io/api/project_badges/measure?project=thomaspoignant_go-feature-flag&metric=alert_status" alt="Sonarcloud Status" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag" alt="License"/></a>
    <a href="https://app.fossa.com/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag"><img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Fthomaspoignant%2Fgo-feature-flag.svg?type=shield" alt="FOSSA Status"/></a>
    <br/>
    <a href="https://github.com/thomaspoignant/go-feature-flag/releases"><img src="https://img.shields.io/github/v/release/thomaspoignant/go-feature-flag" alt="Release version" /></a>
    <a href="https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag"><img src="https://godoc.org/github.com/thomaspoignant/go-feature-flag?status.svg" alt="GoDoc" /></a>
    <img src="https://img.shields.io/github/go-mod/go-version/thomaspoignant/go-feature-flag?logo=go%20version" alt="Go version"/>
    <a href="https://github.com/avelino/awesome-go/#server-applications"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go"></a>
</p>



A feature flag solution, with YAML file in the backend (S3, GitHub, HTTP, local file ...).  
No server to install, just add a file in a central system *(HTTP, S3, GitHub, ...)* and all your services will react to the changes of this file.


If you are not familiar with feature flags also called feature Toggles you can read this [article of Martin Fowler](https://www.martinfowler.com/articles/feature-toggles.html)
that explains why this is a great pattern.  
I've also wrote an [article](https://medium.com/better-programming/feature-flags-and-how-to-iterate-quickly-7e3371b9986) that explains why feature flags can help you to iterate quickly.

## Installation
```bash
go get github.com/thomaspoignant/go-feature-flag
```

## Quickstart
First, you need to initialize the `ffclient` with the location of your backend file.
```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/test.yaml",
    },
})
defer ffclient.Close()
```
*This example will load a file from an HTTP endpoint and will refresh the flags every 3 seconds (if you omit the
PollInterval, the default value is 60 seconds).*

Now you can evaluate your flags anywhere in your code.

```go
user := ffuser.NewUser("user-unique-key")
hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
if hasFlag {
    // flag "test-flag" is true for the user
} else {
    // flag "test-flag" is false for the user
}
```
You can find more example programs in the [examples/](examples) directory.

## Configuration

The configuration is set with `ffclient.Config{}` and you can give it to ``ffclient.Init()`` the initialization
function.

Example:
```go 
ffclient.Init(ffclient.Config{ 
    PollInterval:   3,
    Logger:         log.New(file, "/tmp/log", 0)
    Context         context.Background(),
})
```

|  | Descriptions |
|---|---|
|`PollInterval`   | Number of seconds to wait before refreshing the flags.<br />The default value is 60 seconds.|
|`Logger`   | Logger used to log what `go-feature-flag` is doing.<br />If no logger is provided the module will not log anything.|
|`Context`  | The context used by the retriever.<br />The default value is `context.Background()`.|
|`Retriever`  | The configuration retriever you want to use to get your flag file *(see [Where do I store my flags file](#where-do-i-store-my-flags-file) for the configuration details)*.|

## Where do I store my flags file
`go-feature-flags` support different ways of retrieving the flag file.  
We can have only one source for the file, if you set multiple sources in your configuration, only one will be take in
consideration.

### From GitHub
```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.GithubRetriever{
        RepositorySlug: "thomaspoignant/go-feature-flag",
        Branch: "main",
        FilePath: "testdata/test.yaml",
        GithubToken: "XXXX",
        Timeout: 2 * time.Second,
    },
})
defer ffclient.Close()
```
To configure the access to your GitHub file:
- **RepositorySlug**: your GitHub slug `org/repo-name`. **MANDATORY**
- **FilePath**: the path of your file. **MANDATORY**
- **Branch**: the branch where your file is *(default is `main`)*.
- **GithubToken**: Github token is used to access a private repository, you need the `repo` permission *([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token))*.
- **Timeout**: Timeout for the HTTP call (default is 10 seconds).

:warning: GitHub has rate limits, so be sure to not reach them when setting your `PollInterval`.

### From an HTTP endpoint
```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/test.yaml",
        Timeout: 2 * time.Second,
    },
})
defer ffclient.Close()
```

To configure your HTTP endpoint:
- **URL**: location of your file. **MANDATORY**
- **Method**: the HTTP method you want to use *(default is GET)*.
- **Body**: If you need a body to get the flags.
- **Header**: Header you should pass while calling the endpoint *(useful for authorization)*.
- **Timeout**: Timeout for the HTTP call (default is 10 seconds).

### From a S3 Bucket
```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.S3Retriever{
        Bucket: "tpoi-test",
        Item:   "test.yaml",
        AwsConfig: aws.Config{
            Region: aws.String("eu-west-1"),
        },
    },
})
defer ffclient.Close()
```

To configure your S3 file location:
- **Bucket**: The name of your bucket. **MANDATORY**
- **Item**: The location of your file in the bucket. **MANDATORY**
- **AwsConfig**: An instance of `aws.Config` that configure your access to AWS *(see [this documentation for more info](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html))*. **MANDATORY**

### From a file
```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.FileRetriever{
        Path: "file-example.yaml",
    },
})
defer ffclient.Close()
```

To configure your File retriever:
- **Path**: location of your file. **MANDATORY**

*I will not recommend using a file to store your flags except if it is in a shared folder for all your services.*


## Flags file format
`go-feature-flag` is to avoid to have to host a backend to manage your feature flags and to keep them centralized by using a file a source.  
Your file should be a YAML file with a list of flags *([see example](testdata/test.yaml))*.

A flag configuration looks like:
```yaml
test-flag:
  percentage: 100
  rule: key eq "random-key"
  true: true
  false: false
  default: false
  disable: false
```

|   |   |   |
|---|---|---|
|`test-flag`   |![mandatory](https://img.shields.io/badge/-mandatory-red)   |  Name of the flag. It should be unique.  |
|`percentage`   |![optional](https://img.shields.io/badge/-optional-green)   |  Percentage of the users affect by the flag.<br>**Default value is 0**  |
|`rule`   |![optional](https://img.shields.io/badge/-optional-green)   |  This is the query use to select on which user the flag should apply.<br>Rule format is describe in the [rule format section](#rule-format).<br>**If no rule set, the flag apply to all users *(percentage still apply)*.** |
|`true`   |![mandatory](https://img.shields.io/badge/-mandatory-red)   |  The value return by the flag if apply to the user *(rule is evaluated to true)* and user is in the active percentage. |
|`false`   |![mandatory](https://img.shields.io/badge/-mandatory-red)   |  The value return by the flag if apply to the user *(rule is evaluated to true)* and user is **not** in the active percentage. |
|`default`   |![mandatory](https://img.shields.io/badge/-mandatory-red)   |  The value return by the flag if not apply to the user *(rule is evaluated to false)*. |
|`disable`   |![optional](https://img.shields.io/badge/-optional-green)   |  True if the flag is disabled. |

## Rule format
The rule format is based on the [`nikunjy/rules`](https://github.com/nikunjy/rules) library.

All the operations can be written capitalized or lowercase (ex: eq or EQ can be used).  
Logical Operations supported are `AND` `OR`.

Compare Expression and their definitions (`a|b` means you can use either one of the two `a` or `b`):

```
eq|==: equals to 
ne|!=: not equals to
lt|<: less than 
gt|>: greater than
le|<=: less than equal to
ge|>=: greater than equal to 
co: contains 
sw: starts with 
ew: ends with
in: in a list
pr: present
not: not of a logical expression
```
### Examples

- Select a specific user: `key eq "example@example.com"`
- Select all identified users: `anonymous ne true`
- Select a user with a custom property: `userId eq "12345"`

## Users
Feature flag targeting and rollouts are all determined by the user you pass to your Variation calls.  
The SDK defines a [`User`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#User) struct and a [`UserBuilder`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffuser#UserBuilder) to make this easy. 

Here's an example:

```go
// User with only a key
user1 := ffuser.NewUser("user1-key")

// User with a key plus other attributes
user2 = ffuser.NewUserBuilder("user2-key").
 AddCustom("firstname", "John").
 AddCustom("lastname", "Doe").
 AddCustom("email", "john.doe@example.com").
 Build()
```

The most common attribute is the user's key. In this case we've used the strings "**user1-key**" and "**user2-key**".  
**The user key is the only mandatory user attribute.** The key should also uniquely identify each user. You can use a primary key, an e-mail address, or a hash, as long as the same user always has the same key. We recommend using a hash if possible.

Custom attributes are one of the most powerful features. They let you have rules on these attributes and target users according to any data that you want.

## Variation
The Variation methods determine whether a flag is enabled or not for a specific user.
There is a Variation method for each type: [`BoolVariation`,](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#BoolVariation) [`IntVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#IntVariation), [`Float64Variation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Float64Variation), [`StringVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#StringVariation), [`JSONArrayVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#JSONArrayVariation) and [`JSONVariation`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#JSONVariation).

```go
result, _ := ffclient.BoolVariation("your.feature.key", user, false)

// result is now true or false depending on the setting of this boolean feature flag
```
Variation methods take the feature flag key, a User, and a default value.

The default value is return when an error is encountered _(`ffclient` not initialized, variation with wrong type, flag does not exist ...)._  
In the example, if the flag `your.feature.key` does not exists, result will be `false`.  
Not that you will always have a usable value in the result. 

## Multiple flag configurations
`go-feature-flag` comes ready to use out of the box by calling the `Init` function and after that it will be available everywhere.
Since most applications will want to use a single central flag configuration, the `go-feature-flag` package provides this. It is similar to a singleton.

In all of the examples above, they demonstrate using `go-feature-flag` in its singleton style approach.

### Working with multiple go-feature-flag

You can also create many different `go-feature-flag` client for use in your application.  
Each will have its own unique set of configurations and flags. Each can read from a different config file and from different places.  
All of the functions that `go-feature-flag` package supports are mirrored as methods on a `goFeatureFlag`.

### Example:

```go
x, err := ffclient.New(Config{ Retriever: &ffclient.HTTPRetriever{{URL: "http://example.com/test.yaml",}})
defer x.Close()

y, err := ffclient.New(Config{ Retriever: &ffclient.HTTPRetriever{{URL: "http://example.com/test2.yaml",}})
defer y.Close()

user := ffuser.NewUser("user-key")
x.BoolVariation("test-flag", user, false)
y.BoolVariation("test-flag", user, false)

// ...
```

When working with multiple `go-feature-flag`, it is up to the user to keep track of the different `go-feature-flag` instances.

# How can I contribute?
This project is open for contribution, see the [contributor's guide](CONTRIBUTING.md) for some helpful tips.
