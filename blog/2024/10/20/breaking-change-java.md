# (Java) OpenFeature provider improve the internal cache management.

October 20, 2024 ·

<!-- -->

2 min read

[![Thomas Poignant](https://github.com/thomaspoignant.png)](https://github.com/thomaspoignant)

[Thomas Poignant](https://github.com/thomaspoignant)

Creator of GO Feature Flag

Until Today the java provider, used the guava cache to store the flags and the segments. Using the guava cache is now [discouraged](https://javadoc.io/doc/com.google.guava/guava/latest/com/google/common/cache/package-summary.html) by the guava team.

In order to follow the guidance of the guava team, we have decided to migrate the internal cache of the Java provider from `guava` to `caffeine`.

This may create a **breaking change** for you if you were using a custom cache configuration with the `guava` cache in your provider.<br /><!-- -->Because of this, the cache configuration on `GoFeatureFlagProviderOptions` that used Guava's `CacheBuilder` is now handled by `Caffeine`.

## How to migrate[​](#how-to-migrate "Direct link to How to migrate")

Configuration cache with Guava used to be like this:

```
import com.google.common.cache.CacheBuilder;

// ...

CacheBuilder guavaCacheBuilder = CacheBuilder.newBuilder()

  .initialCapacity(100)

  .maximumSize(2000);



FeatureProvider provider = new GoFeatureFlagProvider(

  GoFeatureFlagProviderOptions

  .builder()

  .endpoint("https://my-gofeatureflag-instance.org")

  .cacheBuilder(guavaCacheBuilder)

  .build());



OpenFeatureAPI.getInstance().setProviderAndWait(provider);



// ...
```

Now with Caffeine it should be like this:

```
import com.github.benmanes.caffeine.cache.Caffeine;

// ...

Caffeine caffeineCacheConfig = Caffeine.newBuilder()

  .initialCapacity(100)

  .maximumSize(2000);



FeatureProvider provider = new GoFeatureFlagProvider(

  GoFeatureFlagProviderOptions

  .builder()

  .endpoint("https://my-gofeatureflag-instance.org")

  .cacheConfig(caffeineCacheConfig)

  .build());



OpenFeatureAPI.getInstance().setProviderAndWait(provider);



// ...
```

For a complete list of customizations options available in Caffeine, please refer to the [Caffeine documentation](https://github.com/ben-manes/caffeine/wiki) for more details.

**Tags:**

* [openfeature](https://gofeatureflag.org/blog/tags/openfeature)
* [java](https://gofeatureflag.org/blog/tags/java)
* [breaking changes](https://gofeatureflag.org/blog/tags/breaking-changes)

[Edit this page](https://github.com/thomaspoignant/go-feature-flag/tree/main/website/blog/2024-10-20-breaking-change-java/index.md)
