export const tsSnippet = `import {OpenFeature} from '@openfeature/web-sdk';
import {GoFeatureFlagWebProvider} from '@openfeature/go-feature-flag-web-provider';

await OpenFeature.setProviderAndWait(
  new GoFeatureFlagWebProvider({endpoint: 'http://localhost:1031'})
);
await OpenFeature.setContext({targetingKey: 'user-123'});
OpenFeature.setProvider(goFeatureFlagWebProvider);

const client = await OpenFeature.getClient();
const enabled = await client.getBooleanValue('my-new-feature', false);`;

export const goSnippet = `import (
    "context"
    of "github.com/open-feature/go-sdk/openfeature"
    gofeatureflag "github.com/open-feature/go-sdk-contrib/providers/go-feature-flag"
)

provider, _ := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{
    Endpoint: "http://localhost:1031",
})
of.SetProvider(provider)
client := of.NewClient("my-app")
enabled, _ := client.BooleanValue(context.Background(), "my-new-feature", false,
    of.NewEvaluationContext("user-123", nil))`;

export const javaSnippet = `import dev.openfeature.sdk.*;
import dev.openfeature.contrib.providers.gofeatureflag.*;

FeatureProvider provider = new GoFeatureFlagProvider(
        GoFeatureFlagProviderOptions.builder()
                .endpoint("http://localhost:1031")
                .build());

OpenFeatureAPI.getInstance().setProviderAndWait(provider);
Client client = OpenFeatureAPI.getInstance().getClient();
boolean enabled = client.getBooleanValue("my-new-feature", false,
    new MutableContext("user-123"));`;

export const dotnetSnippet = `using OpenFeature;
using OpenFeature.Providers.GOFeatureFlag;

var options = new GoFeatureFlagProviderOptions { Endpoint = "https://my-instance.gofeatureflag.org" };
var provider = new GoFeatureFlagProvider(options);

await Api.Instance.SetProviderAsync("my-app", provider);
var client = Api.Instance.GetClient("my-app");

var evaluationContext = EvaluationContext.Builder()
  .SetTargetingKey("user-123")
  .Build();

var enabled = await client.GetBooleanValueAsync("my-new-feature", false, evaluationContext);`;

export const swiftSnippet = `import GOFeatureFlag
import OpenFeature

let options = GoFeatureFlagProviderOptions(endpoint: "http://localhost:1031")
let provider = GoFeatureFlagProvider(options: options)

let evaluationContext = MutableContext(targetingKey: "user-123", structure: MutableStructure())
OpenFeatureAPI.shared.setProvider(provider: provider, initialContext: evaluationContext)

let client = OpenFeatureAPI.shared.getClient()
let enabled = await client.getBooleanValueAsync(key: "my-new-feature", defaultValue: false, evaluationContext: evaluationContext)`;

export const kotlinSnippet = `import dev.openfeature.kotlin-sdk.*
import org.gofeatureflag.openfeature.*

val evaluationContext: EvaluationContext = ImmutableContext(
        targetingKey = "user-123"
    )

OpenFeatureAPI.setProvider(
    GoFeatureFlagProvider(
        options = GoFeatureFlagOptions(endpoint = "http://localhost:1031")
    ),
    evaluationContext
)
    
val client = OpenFeatureAPI.getClient()
val enabled = client.getBooleanValue("my-new-feature", false)`;
