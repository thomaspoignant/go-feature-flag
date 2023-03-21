package org.gofeatureflag.integrationtests;


import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProvider;
import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProviderOptions;
import dev.openfeature.contrib.providers.gofeatureflag.exception.InvalidOptions;
import dev.openfeature.sdk.*;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.Arrays;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class ProviderTests{
    private static final String relayProxyEndpoint = "http://localhost:1031/";
//    @DisplayName("Single test successful")
//    @Test
//    void testSingleSuccessTest() throws InvalidOptions {
//        GoFeatureFlagProviderOptions options = GoFeatureFlagProviderOptions.builder().endpoint(relayProxyEndpoint).build();
//        GoFeatureFlagProvider provider = new GoFeatureFlagProvider(options);
//
//        OpenFeatureAPI.getInstance().setProvider(provider);
//        OpenFeatureAPI api = OpenFeatureAPI.getInstance();
//        Client featureFlagClient = api.getClient();
//
//    }

    @DisplayName("bool: should resolve a valid boolean flag with TARGETING_MATCH reason")
    @Test
    void boolShouldResolveAValidBooleanFlagWithTargetingMatchReason() throws InvalidOptions {
        GoFeatureFlagProviderOptions options = GoFeatureFlagProviderOptions.builder().endpoint(relayProxyEndpoint).build();
        GoFeatureFlagProvider provider = new GoFeatureFlagProvider(options);
        OpenFeatureAPI.getInstance().setProvider(provider);
        OpenFeatureAPI api = OpenFeatureAPI.getInstance();
        Client featureFlagClient = api.getClient();
        MutableContext userContext = new MutableContext()
        .add("email",        "john.doe@gofeatureflag.org")
        .add("firstname",    "john")
        .add("lastname",     "doe")
        .add("anonymous",    false)
        .add("professional", true)
        .add("rate",         3.14)
        .add("age",          30)
        .add("admin",        true)
        .add("labels", Arrays.asList(new Value("pro"), new Value("beta")))
        .add("company_info",  new MutableStructure().add("name", "my_company").add("size", 120));
        userContext.setTargetingKey("d45e303a-38c2-11ed-a261-0242ac120002");


        FlagEvaluationDetails<Boolean> res = featureFlagClient.getBooleanDetails("bool_targeting_match", false, userContext);
        assertEquals(Reason.TARGETING_MATCH.toString(), res.getReason());

//        Boolean adminFlag = featureFlagClient.getBooleanValue("flag-only-for-admin", false, userContext);
//        if (adminFlag) {
//            // flag "flag-only-for-admin" is true for the user
//        } else {
//            // flag "flag-only-for-admin" is false for the user
//        }
    }
}

