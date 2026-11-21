# 🏁 Getting Started

## Installation[​](#installation "Direct link to Installation")

```
go get github.com/thomaspoignant/go-feature-flag
```

## Create a feature flag configuration[​](#create-a-feature-flag-configuration "Direct link to Create a feature flag configuration")

Create a new `YAML` file containing your first flag configuration.

flag-config.goff.yaml

```
# 20% of the users will use the variation "my-new-feature"

test-flag:

  variations:

    my-new-feature: true

    my-old-feature: false

  defaultRule:

    percentage:

      my-new-feature: 20

      my-old-feature: 80
```

This flag split the usage of this flag, 20% will use the variation `my-new-feature` and 80% the variation `my-old-feature`.

## SDK Initialisation[​](#sdk-initialisation "Direct link to SDK Initialisation")

First, you need to initialize the `ffclient` with the location of your backend file.

```
err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever:      &fileretriever.Retriever{

        Path: "flag-config.goff.yaml",

    },

})

defer ffclient.Close()
```

*This example will load a file from your local computer and will refresh the flags every 3 seconds (if you omit the PollingInterval, the default value is 60 seconds).*

tip

This is a basic configuration to test locally, in production it is better to use a remote place to store your feature flag configuration file.

Look at the list of available options in the [**Store your feature flag file** page](https://gofeatureflag.org/docs/go_module/store_file.md).

## Evaluate your flags[​](#evaluate-your-flags "Direct link to Evaluate your flags")

Now you can evaluate your flags anywhere in your code.

```
user := ffcontext.NewEvaluationContext("user-unique-key")

hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)

if hasFlag {

    // flag "test-flag" is true for the user

} else {

    // flag "test-flag" is false for the user

}
```

You can find more examples in the [examples/](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples) directory.
