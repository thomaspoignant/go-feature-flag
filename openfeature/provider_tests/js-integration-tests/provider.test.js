const {describe, expect, it, beforeEach} = require('@jest/globals');
const {OpenFeature} = require("@openfeature/js-sdk");
const {GoFeatureFlagProvider} = require("@openfeature/go-feature-flag-provider");
describe('Provider tests', () => {
  let goffClient;
  let userCtx;

  beforeEach(() => {
    // init Open Feature SDK with GO Feature Flag provider
    const goFeatureFlagProvider = new GoFeatureFlagProvider({
      endpoint: 'http://localhost:1031/' // DNS of your instance of relay proxy
    });
    goffClient = OpenFeature.getClient('my-app')
    OpenFeature.setProvider(goFeatureFlagProvider);

    userCtx = {
      targetingKey: 'd45e303a-38c2-11ed-a261-0242ac120002', // user unique identifier (mandatory)
      firstname: 'john',
      lastname: 'doe',
      email: 'john.doe@gofeatureflag.org',
      anonymous: false,
      professional: true,
      rate: 3.14,
      age: 30,
      admin: true,
      labels: ["pro", "beta"],
      company_info: {name: "my_company", "size": 120}
    };
  });

  describe("bool", () => {
    it('should resolve a valid boolean flag with TARGETING_MATCH reason', async () => {
      const flagKey = "bool_targeting_match"
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: true,
        variant: "True",
      }
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx)
      expect(expected).toEqual(got)
    })

    it('should use boolean default value if the flag is disabled', async () => {
      const flagKey = "disabled_bool"
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        value: false,
      }
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx)
      expect(expected).toEqual(got)
    })

    it('should use boolean default value if the flag is disabled', async () => {
      const flagKey = "string_key"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "TYPE_MISMATCH",
        errorMessage: "Flag value string_key had unexpected type string, expected boolean."
      }
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx)
      expect(expected).toEqual(got)
    })

    it('should error if flag does not exists', async () => {
      const flagKey = "does_not_exists"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "FLAG_NOT_FOUND",
        errorMessage: "Flag does_not_exists was not found in your configuration"
      }
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx)
      expect(expected).toEqual(got)
    })
  })
  describe("string", () => {
    it('should resolve a valid string flag with TARGETING_MATCH reason', async () => {
      const flagKey = "string_key"
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: "CC0000",
        variant: "True",
      }
      const got = await goffClient.getStringDetails(flagKey, "default", userCtx)
      expect(expected).toEqual(got)
    })

    it('should use string default value if the flag is disabled', async () => {
      const flagKey = "disabled_string"
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        value: "default",
      }
      const got = await goffClient.getStringDetails(flagKey, "default", userCtx)
      expect(expected).toEqual(got)
    })

    it('should error if we expect a string and got another type', async () => {
      const flagKey = "bool_targeting_match"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "TYPE_MISMATCH",
        value: "default",
        errorMessage: "Flag value bool_targeting_match had unexpected type boolean, expected string."
      }
      const got = await goffClient.getStringDetails(flagKey, "default", userCtx)
      expect(expected).toEqual(got)
    })

    it('should error if flag does not exists', async () => {
      const flagKey = "does_not_exists"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "FLAG_NOT_FOUND",
        value: "default",
        errorMessage: "Flag does_not_exists was not found in your configuration"
      }
      const got = await goffClient.getStringDetails(flagKey, "default", userCtx)
      expect(expected).toEqual(got)
    })
  })
  describe("number", () => {
    it('should resolve a valid number flag with TARGETING_MATCH reason', async () => {
      const flagKey = "double_key"
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: 100.25,
        variant: "True",
      }
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx)
      expect(expected).toEqual(got)
    })

    it('should resolve a valid number flag with TARGETING_MATCH reason', async () => {
      const flagKey = "disabled_float"
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        value: 123.45,
      }
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx)
      expect(expected).toEqual(got)
    })

    it('should error if we expect a number and got another type', async () => {
      const flagKey = "bool_targeting_match"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "TYPE_MISMATCH",
        errorMessage: "Flag value bool_targeting_match had unexpected type boolean, expected number.",
        value: 123.45,
      }
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx)
      expect(expected).toEqual(got)
    })

    it('should error if flag does not exists', async () => {
      const flagKey = "does_not_exists"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "FLAG_NOT_FOUND",
        errorMessage: "Flag does_not_exists was not found in your configuration",
        value: 123.45,
      }
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx)
      expect(expected).toEqual(got)
    })
  })
  describe("object", () => {
    it('should resolve a valid object flag with TARGETING_MATCH reason', async () => {
      const flagKey = "object_key"
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: {
          test4: 1,
          test2: false,
          test3: 123.3,
          test:"test1"
        },
        variant: "True",
      }
      const got = await goffClient.getObjectDetails(flagKey, {"test":1234}, userCtx)
      expect(expected).toEqual(got)
    })

    it('should use object default value if the flag is disabled', async () => {
      const flagKey = "disabled_int"
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        value: {"test":1234},
      }
      const got = await goffClient.getObjectDetails(flagKey, {"test":1234}, userCtx)
      expect(expected).toEqual(got)
    })

    it('should error if we expect an object and got another type', async () => {
      const flagKey = "bool_targeting_match"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "TYPE_MISMATCH",
        errorMessage: "Flag value bool_targeting_match had unexpected type boolean, expected object.",
        value: {"test":1234},
      }
      const got = await goffClient.getObjectDetails(flagKey, {"test":1234}, userCtx)
      expect(expected).toEqual(got)
    })

    it('should error if flag does not exists', async () => {
      const flagKey = "does_not_exists"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "FLAG_NOT_FOUND",
        errorMessage: "Flag does_not_exists was not found in your configuration",
        value: {"test":1234},
      }
      const got = await goffClient.getObjectDetails(flagKey, {"test":1234}, userCtx)
      expect(expected).toEqual(got)
    })
  })
  describe("authenticated relay proxy", () => {
    it('should resolve a valid flag with an apiKey', async () => {
      const goFeatureFlagProvider = new GoFeatureFlagProvider({
        endpoint: 'http://localhost:1032/',
        apiKey: "authorized_token"
      });
      goffClient = OpenFeature.getClient('my-app')
      OpenFeature.setProvider(goFeatureFlagProvider);

      const flagKey = "bool_targeting_match"
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: true,
        variant: "True",
      }
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx)
      expect(expected).toEqual(got)
    });

    it('should resolve a default value with an invalid apiKey', async () => {
      const goFeatureFlagProvider = new GoFeatureFlagProvider({
        endpoint: 'http://localhost:1032/',
        apiKey: "invalid-api-key"
      });
      goffClient = OpenFeature.getClient('my-app')
      OpenFeature.setProvider(goFeatureFlagProvider);

      const flagKey = "bool_targeting_match"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "GENERAL",
        errorMessage: "invalid token used to contact GO Feature Flag relay proxy instance"
      }
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx)
      expect(expected).toEqual(got)
    });

    it('should resolve a default value with an empty apiKey', async () => {
      const goFeatureFlagProvider = new GoFeatureFlagProvider({
        endpoint: 'http://localhost:1032/',
        apiKey: ""
      });
      goffClient = OpenFeature.getClient('my-app')
      OpenFeature.setProvider(goFeatureFlagProvider);

      const flagKey = "bool_targeting_match"
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "GENERAL",
        errorMessage: "invalid token used to contact GO Feature Flag relay proxy instance"
      }
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx)
      expect(expected).toEqual(got)
    });
  });
});
