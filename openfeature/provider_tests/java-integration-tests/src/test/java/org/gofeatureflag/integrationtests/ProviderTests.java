package org.gofeatureflag.integrationtests;


import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProvider;
import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProviderOptions;
import dev.openfeature.contrib.providers.gofeatureflag.exception.InvalidOptions;
import dev.openfeature.sdk.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.Arrays;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class ProviderTests {
    private static final String relayProxyEndpoint = "http://localhost:1031";
    private EvaluationContext defaultEvaluationContext;
    private Client goffClient;

    @BeforeEach
    void init() throws InvalidOptions {
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
    }


    @DisplayName("bool: should resolve a valid boolean flag with TARGETING_MATCH reason")
    @Test
    void boolShouldResolveAValidBooleanFlagWithTargetingMatchReason() {
        String flagKey = "bool_targeting_match";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.TARGETING_MATCH.toString())
                .value(true)
                .variant("True")
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails(flagKey, false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("bool: should use boolean default value if the flag is disabled")
    @Test
    void boolShouldUseBooleanDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_bool";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(false)
                .build();
        FlagEvaluationDetails<Boolean> got = goffClient.getBooleanDetails(flagKey, false, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("bool: should error if we expect a boolean and got another type")
    @Test
    void boolShouldErrorIfWeExpectABooleanAndGotAnotherType() {
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
    void stringShouldResolveAValidStringFlagWithTargetingMatchReason() {
        String flagKey = "string_key";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value("CC0000")
                .variant("True")
                .build();
        FlagEvaluationDetails<String> got = goffClient.getStringDetails(flagKey, "default", defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("string: should use string default value if the flag is disabled")
    @Test
    void stringShouldUseStringDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_string";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value("default")
                .build();
        FlagEvaluationDetails<String> got = goffClient.getStringDetails(flagKey, "default", defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("string: should error if we expect a string and got another type")
    @Test
    void stringShouldErrorIfWeExpectAStringAndGotAnotherType() {
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

    @DisplayName("double: should resolve a valid string flag with TARGETING_MATCH reason")
    @Test
    void doubleShouldResolveAValidStringFlagWithTargetingMatchReason() {
        String flagKey = "double_key";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value(100.25)
                .variant("True")
                .build();
        FlagEvaluationDetails<Double> got = goffClient.getDoubleDetails(flagKey, 123.45, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("double: should use string default value if the flag is disabled")
    @Test
    void doubleShouldUseStringDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_float";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(123.45)
                .build();
        FlagEvaluationDetails<Double> got = goffClient.getDoubleDetails(flagKey, 123.45, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("double: should error if we expect a string and got another type")
    @Test
    void doubleShouldErrorIfWeExpectAStringAndGotAnotherType() {
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

    @DisplayName("int: should resolve a valid string flag with TARGETING_MATCH reason")
    @Test
    void intShouldResolveAValidStringFlagWithTargetingMatchReason() {
        String flagKey = "integer_key";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value(100)
                .variant("True")
                .build();
        FlagEvaluationDetails<Integer> got = goffClient.getIntegerDetails(flagKey, 123, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("int: should use string default value if the flag is disabled")
    @Test
    void intShouldUseStringDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_int";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(123)
                .build();
        FlagEvaluationDetails<Integer> got = goffClient.getIntegerDetails(flagKey, 123, defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("int: should error if we expect a string and got another type")
    @Test
    void intShouldErrorIfWeExpectAStringAndGotAnotherType() {
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
    void objectShouldResolveAValidObjectFlagWithTargetingMatchReason() {
        String flagKey = "object_key";
        Structure expectedValue = new MutableStructure().add("test4", 1).add("test2", false).add("test3", 123.3).add("test", "test1");
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .reason(Reason.TARGETING_MATCH.toString())
                .flagKey(flagKey)
                .value(new Value(expectedValue))
                .variant("True")
                .build();
        FlagEvaluationDetails<Value> got = goffClient.getObjectDetails(flagKey, new Value(123), defaultEvaluationContext);
        assertEquals(expected, got);
    }

    @DisplayName("object: should use object default value if the flag is disabled")
    @Test
    void objectShouldUseStringDefaultValueIfTheFlagIsDisabled() {
        String flagKey = "disabled_int";
        FlagEvaluationDetails expected = FlagEvaluationDetails.builder()
                .flagKey(flagKey)
                .reason(Reason.DISABLED.toString())
                .value(new Value(123))
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
}

