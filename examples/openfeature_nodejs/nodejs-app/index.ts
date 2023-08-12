import { GoFeatureFlagProvider } from "@openfeature/go-feature-flag-provider";
import { EvaluationContext, OpenFeature } from "@openfeature/js-sdk";

async function main(contexts: Array<EvaluationContext>) {
  // We start by creating an instance of the Go Feature Flag Provider
  // We are just setting the endpoint to connect to our instance of GO Feature Flag
  const provider = new GoFeatureFlagProvider({
    endpoint: "http://gofeatureflag:1031",
    disableDataCollection: true,
    flagCacheSize: 1000, 
    flagCacheTTL: 10000 // we keep the value in the cache for 10 seconds
  });

  // We associate the provider to the SDK
  // It means that now when we call OpenFeature it will rely on GO Feature Flag as a backend.
  OpenFeature.setProvider(provider);

  // We need to ask for a client to evaluate the flags.
  const client = OpenFeature.getClient();

  let it = 0;
  while (true) {
    const ctx = contexts[it % contexts.length];

    const flag = "my-new-feature";
    const myNewFeature = await client.getBooleanValue(flag, false, ctx);
    if (myNewFeature) {
      // the new feature is available
      console.log(`✅ ${flag} is available for id ${ctx.targetingKey}`);
    } else {
      // apply the old feature
      console.log(`❌ ${flag} is not available for id ${ctx.targetingKey}`);
    }

    await new Promise(resolve => setTimeout(resolve, 1000));
    it++;
  }
}

const contexts = [
  { targetingKey: "aae1cb41-c3cb-4753-a117-031ddc958f82", email: "user1@gofeatureflag.org" },
  { targetingKey: "fa0f8cfa-02a8-4424-b201-f4dca70a3819", email: "user2@gofeatureflag.org" },
  { targetingKey: "401ee3dd-81f0-4f49-9bf0-a95eb1f1d0d6", email: "user3@gofeatureflag.org" },
  { targetingKey: "9799dce2-9621-4137-8a95-6033cdeeddc5", email: "user4@gofeatureflag.org" },
  { targetingKey: "628836b5-8d64-4ba5-8043-11e1cf811e58", email: "user5@gofeatureflag.org" },
  { targetingKey: "b3576da1-1b5b-4b94-b98c-ac3eb92dd53f", email: "user6@gofeatureflag.org" }
];

main(contexts);