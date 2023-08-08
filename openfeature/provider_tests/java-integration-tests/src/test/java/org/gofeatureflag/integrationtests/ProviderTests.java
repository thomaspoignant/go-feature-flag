package org.gofeatureflag.integrationtests;


import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProvider;
import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProviderOptions;
import dev.openfeature.contrib.providers.gofeatureflag.exception.InvalidOptions;
import dev.openfeature.sdk.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.Arrays;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;
import java.util.function.Consumer;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class ProviderTests {
    private static final String relayProxyEndpoint = "http://localhost:1031";
    private static final String relayProxyAuthenticatedEndpoint = "http://localhost:1032";
    private EvaluationContext defaultEvaluationContext;
    private Client goffClient;

    private ImmutableMetadata defaultMetadata = ImmutableMetadata.builder()
            .addString("description", "this is a test")
            .addString("pr_link", "https://github.com/thomaspoignant/go-feature-flag/pull/916")
            .build();
    @BeforeEach
    void init() throws InvalidOptions, ExecutionException, InterruptedException {
        MutableContext userContext = new MutableContext()
                .add("email", "john.doe@gofeatureflag.org")
                .add("firstname", "john")
                .add("lastname", "doe")
                .add("anonymous", false)
                .add("professional", true)
                .add("rate", 3.14)
                .add("age", 30)
                .add("admin", true)
                .add("labels", Arrays.asList(new Value("pro"), new Value("beta")))
                .add("company_info", new MutableStructure().add("name", "my_company").add("size", 120));
        userContext.setTargetingKey("d45e303a-38c2-11ed-a261-0242ac120002");
        defaultEvaluationContext = userContext;

        GoFeatureFlagProviderOptions options = GoFeatureFlagProviderOptions.builder().endpoint(relayProxyEndpoint).build();
        GoFeatureFlagProvider provider = new GoFeatureFlagProvider(options);
        OpenFeatureAPI.getInstance().setProvider(provider);
        OpenFeatureAPI api = OpenFeatureAPI.getInstance();
        goffClient = api.getClient();
        this.waitProviderReady();
    }


    @DisplayName("bool: should resolve a valid boolean flag with TARGETING_MATCH reason")
    @Test
    void shouldResolveAValidBooleanFlagWithTargetingMatchReason() {
        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.TARGETING_MATCH.toString())
                .value(true)
                .variant("True")
                .flagMetadata(defaultMetadata)
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails(flagKey, false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("bool: should use boolean default value if the flag is disabled")
    @Test
    void shouldUseBooleanDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_bool";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(false)
                .variant("SdkDefault")
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails(flagKey, false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("bool: should error if we expect a boolean and got another type")
    @Test
    void shouldErrorIfWeExpectABooleanAndGotAnotherType() {
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(false)
                .errorCode(ErrorCode.TYPE_MISMATCH)
                .errorMessage("Flag value string_key had unexpected type class java.lang.String, expected class java.lang.Boolean.")
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails("string_key", false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("bool: should error if flag does not exists")
    @Test
    void boolShouldErrorIfFlagDoesNotExists() {
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(false)
                .errorCode(ErrorCode.FLAG_NOT_FOUND)
                .errorMessage("Flag does_not_exists was not found in your configuration")
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails("does_not_exists", false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("string: should resolve a valid string flag with TARGETING_MATCH reason")
    @Test
    void shouldResolveAValidStringFlagWithTargetingMatchReason() {
        String flagKey = "string_key";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value("CC0000")
                .variant("True")
                .flagMetadata(defaultMetadata)
                .build();
        FlagEvaluationDetails<String> got = goffClient.getStringDetails(flagKey, "default", defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("string: should use string default value if the flag is disabled")
    @Test
    void shouldUseStringDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_string";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value("default")
                .variant("SdkDefault")
                .build();
        FlagEvaluationDetails<String> got = goffClient.getStringDetails(flagKey, "default", defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("string: should error if we expect a string and got another type")
    @Test
    void shouldErrorIfWeExpectAStringAndGotAnotherType() {
        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value("default")
                .errorCode(ErrorCode.TYPE_MISMATCH)
                .errorMessage("Flag value bool_targeting_match had unexpected type class java.lang.Boolean, expected class java.lang.String.")
                .build();
        FlagEvaluationDetails<String> got = goffClient.getStringDetails(flagKey, "default", defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("string: should error if flag does not exists")
    @Test
    void stringShouldErrorIfFlagDoesNotExists() {
        String flagKey = "does_not_exists";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value("default")
                .errorCode(ErrorCode.FLAG_NOT_FOUND)
                .errorMessage("Flag does_not_exists was not found in your configuration")
                .build();
        FlagEvaluationDetails<String> got = goffClient.getStringDetails(flagKey, "default", defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("double: should resolve a valid double flag with TARGETING_MATCH reason")
    @Test
    void shouldResolveAValidDoubleFlagWithTargetingMatchReason() {
        String flagKey = "double_key";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value(100.25)
                .variant("True")
                .flagMetadata(defaultMetadata)
                .build();
        FlagEvaluationDetails<Double> got = goffClient.getDoubleDetails(flagKey, 123.45, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("double: should use double default value if the flag is disabled")
    @Test
    void shouldUseDoubleDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_float";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(123.45)
                .variant("SdkDefault")
                .build();
        FlagEvaluationDetails<Double> got = goffClient.getDoubleDetails(flagKey, 123.45, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("double: should error if we expect a double and got another type")
    @Test
    void shouldErrorIfWeExpectADoubleAndGotAnotherType() {
        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(123.45)
                .errorCode(ErrorCode.TYPE_MISMATCH)
                .errorMessage("Flag value bool_targeting_match had unexpected type class java.lang.Boolean, expected class java.lang.Double.")
                .build();
        FlagEvaluationDetails<Double> got = goffClient.getDoubleDetails(flagKey, 123.45, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("double: should error if flag does not exists")
    @Test
    void doubleShouldErrorIfFlagDoesNotExists() {
        String flagKey = "does_not_exists";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(123.45)
                .errorCode(ErrorCode.FLAG_NOT_FOUND)
                .errorMessage("Flag does_not_exists was not found in your configuration")
                .build();
        FlagEvaluationDetails<Double> got = goffClient.getDoubleDetails(flagKey, 123.45, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("int: should resolve a valid integer flag with TARGETING_MATCH reason")
    @Test
    void shouldResolveAValidIntFlagWithTargetingMatchReason() {
        String flagKey = "integer_key";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value(100)
                .variant("True")
                .flagMetadata(defaultMetadata)
                .build();
        FlagEvaluationDetails<Integer> got = goffClient.getIntegerDetails(flagKey, 123, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("int: should use integer default value if the flag is disabled")
    @Test
    void shouldUseIntDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_int";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(123)
                .variant("SdkDefault")
                .build();
        FlagEvaluationDetails<Integer> got = goffClient.getIntegerDetails(flagKey, 123, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("int: should error if we expect a integer and got another type")
    @Test
    void shouldErrorIfWeExpectAIntAndGotAnotherType() {
        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(123)
                .errorCode(ErrorCode.TYPE_MISMATCH)
                .errorMessage("Flag value bool_targeting_match had unexpected type class java.lang.Boolean, expected class java.lang.Integer.")
                .build();
        FlagEvaluationDetails<Integer> got = goffClient.getIntegerDetails(flagKey, 123, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("int: should error if flag does not exists")
    @Test
    void intShouldErrorIfFlagDoesNotExists() {
        String flagKey = "does_not_exists";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(123)
                .errorCode(ErrorCode.FLAG_NOT_FOUND)
                .errorMessage("Flag does_not_exists was not found in your configuration")
                .build();
        FlagEvaluationDetails<Integer> got = goffClient.getIntegerDetails(flagKey, 123, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("object: should resolve a valid object flag with TARGETING_MATCH reason")
    @Test
    void shouldResolveAValidObjectFlagWithTargetingMatchReason() {
        String flagKey = "object_key";
        Structure expectedValue = new MutableStructure().add("test4", 1).add("test2", false).add("test3", 123.3).add("test", "test1");
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value(new Value(expectedValue))
                .variant("True")
                .flagMetadata(defaultMetadata)
                .build();
        FlagEvaluationDetails<Value> got = goffClient.getObjectDetails(flagKey, new Value(123), defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("object: should use object default value if the flag is disabled")
    @Test
    void shouldUseObjectDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_int";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(new Value(123))
                .variant("SdkDefault")
                .build();
        FlagEvaluationDetails<Value> got = goffClient.getObjectDetails(flagKey, new Value(123), defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("object: should error if flag does not exists")
    @Test
    void objectShouldErrorIfFlagDoesNotExists() {
        String flagKey = "does_not_exists";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .errorCode(ErrorCode.FLAG_NOT_FOUND)
                .errorMessage("Flag does_not_exists was not found in your configuration")
                .value(new Value(123))
                .build();
        FlagEvaluationDetails<Value> got = goffClient.getObjectDetails(flagKey, new Value(123), defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("authenticated relay proxy: valid")
    @Test
    void authenticatedRelayProxyValid() throws InvalidOptions, ExecutionException, InterruptedException {
        GoFeatureFlagProviderOptions options = GoFeatureFlagProviderOptions.builder()
                .apiKey("authorized_token").endpoint(relayProxyAuthenticatedEndpoint).build();
        GoFeatureFlagProvider provider = new GoFeatureFlagProvider(options);
        OpenFeatureAPI.getInstance().setProvider(provider);
        OpenFeatureAPI api = OpenFeatureAPI.getInstance();
        goffClient = api.getClient();
        this.waitProviderReady();

        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.TARGETING_MATCH.toString())
                .value(true)
                .variant("True")
                .flagMetadata(defaultMetadata)
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails(flagKey, false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("authenticated relay proxy: empty api key")
    @Test
    void authenticatedRelayProxyEmptyToken() throws InvalidOptions, ExecutionException, InterruptedException {
        GoFeatureFlagProviderOptions options = GoFeatureFlagProviderOptions.builder()
                .apiKey("").endpoint(relayProxyAuthenticatedEndpoint).build();
        GoFeatureFlagProvider provider = new GoFeatureFlagProvider(options);
        OpenFeatureAPI.getInstance().setProvider(provider);
        OpenFeatureAPI api = OpenFeatureAPI.getInstance();
        goffClient = api.getClient();
        this.waitProviderReady();

        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(false)
                .errorCode(ErrorCode.GENERAL)
                .errorMessage("impossible to contact GO Feature Flag relay proxy instance")
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails(flagKey, false, defaultEvaluationContext);
        assertEquals(expected, got);

    }

    @DisplayName("authenticated relay proxy: invalid api key")
    @Test
    void authenticatedRelayProxyInvalidToken() throws InvalidOptions, ExecutionException, InterruptedException {
        GoFeatureFlagProviderOptions options = GoFeatureFlagProviderOptions.builder()
                .apiKey("invalid-api-key").endpoint(relayProxyAuthenticatedEndpoint).build();
        GoFeatureFlagProvider provider = new GoFeatureFlagProvider(options);
        OpenFeatureAPI.getInstance().setProvider(provider);
        OpenFeatureAPI api = OpenFeatureAPI.getInstance();
        goffClient = api.getClient();
        this.waitProviderReady();

        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.ERROR.toString())
                .value(false)
                .errorCode(ErrorCode.GENERAL)
                .errorMessage("invalid token used to contact GO Feature Flag relay proxy instance")
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails(flagKey, false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    private void waitProviderReady() throws ExecutionException, InterruptedException {
        CompletableFuture<EventDetails> completableFuture = new CompletableFuture<>();
        OpenFeatureAPI.getInstance().onProviderReady(new Consumer<EventDetails>() {
            @Override
            public void accept(EventDetails eventDetails) {
                completableFuture.complete(eventDetails);
            }
        });
        completableFuture.get();
    }
}



