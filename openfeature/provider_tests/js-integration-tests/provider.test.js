const {
  describe,
  expect,
  it,
  beforeEach,
  afterEach,
} = require("@jest/globals");
const { OpenFeature } = require("@openfeature/server-sdk");
const {
  GoFeatureFlagProvider,
  UnauthorizedError,
} = require("@openfeature/go-feature-flag-provider");
describe("Provider tests", () => {
  let goffClient;
  let userCtx;

  beforeEach(async () => {
    // init Open Feature SDK with GO Feature Flag provider
    const goFeatureFlagProvider = new GoFeatureFlagProvider({
      endpoint: "http://localhost:1031/", // DNS of your instance of relay proxy
    });
    goffClient = OpenFeature.getClient("my-app");
    await OpenFeature.setProviderAndWait("my-app", goFeatureFlagProvider);
    OpenFeature.setContext({
      gofeatureflag: {
        flagList: ["flag1", "flag2"],
      },
    });

    userCtx = {
      targetingKey: "d45e303a-38c2-11ed-a261-0242ac120002", // user unique identifier (mandatory)
      firstname: "john",
      lastname: "doe",
      email: "john.doe@gofeatureflag.org",
      anonymous: false,
      professional: true,
      rate: 3.14,
      age: 30,
      admin: true,
      labels: ["pro", "beta"],
      company_info: { name: "my_company", size: 120 },
    };
  });

  afterEach(async () => {
    await OpenFeature.close();
  });

  describe("bool", () => {
    it("should resolve a valid boolean flag with TARGETING_MATCH reason", async () => {
      const flagKey = "bool_targeting_match";
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: true,
        variant: "True",
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
      };
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx);
      expect(got).toEqual(expected);
    });

    it("should use boolean default value if the flag is disabled", async () => {
      const flagKey = "disabled_bool";
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
        variant: "SdkDefault",
        value: false,
      };
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx);
      expect(got).toEqual(expected);
    });

    it("should error if type_mismatch", async () => {
      const flagKey = "string_key";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "TYPE_MISMATCH",
        flagMetadata: {},
        errorMessage: "Flag string_key had unexpected type, expected boolean.",
      };
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx);
      expect(got).toEqual(expected);
    });

    it("should error if flag does not exists", async () => {
      const flagKey = "does_not_exists";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "FLAG_NOT_FOUND",
        flagMetadata: {},
        errorMessage: "Flag with key 'does_not_exists' not found",
      };
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx);
      expect(got).toEqual(expected);
    });
  });
  describe("string", () => {
    it("should resolve a valid string flag with TARGETING_MATCH reason", async () => {
      const flagKey = "string_key";
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: "CC0000",
        variant: "True",
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
      };
      const got = await goffClient.getStringDetails(
        flagKey,
        "default",
        userCtx
      );
      expect(got).toEqual(expected);
    });

    it("should use string default value if the flag is disabled", async () => {
      const flagKey = "disabled_string";
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        value: "default",
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
        variant: "SdkDefault",
      };
      const got = await goffClient.getStringDetails(
        flagKey,
        "default",
        userCtx
      );
      expect(got).toEqual(expected);
    });

    it("should error if we expect a string and got another type", async () => {
      const flagKey = "bool_targeting_match";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "TYPE_MISMATCH",
        value: "default",
        flagMetadata: {},
        errorMessage:
          "Flag bool_targeting_match had unexpected type, expected string.",
      };
      const got = await goffClient.getStringDetails(
        flagKey,
        "default",
        userCtx
      );
      expect(got).toEqual(expected);
    });

    it("should error if flag does not exists", async () => {
      const flagKey = "does_not_exists";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "FLAG_NOT_FOUND",
        value: "default",
        flagMetadata: {},
        errorMessage: "Flag with key 'does_not_exists' not found",
      };
      const got = await goffClient.getStringDetails(
        flagKey,
        "default",
        userCtx
      );
      expect(got).toEqual(expected);
    });
  });
  describe("number", () => {
    it("should resolve a valid number flag with TARGETING_MATCH reason", async () => {
      const flagKey = "double_key";
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: 100.25,
        variant: "True",
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
      };
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx);
      expect(got).toEqual(expected);
    });

    it("should use number default value if the flag is disabled", async () => {
      const flagKey = "disabled_float";
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        value: 123.45,
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
        variant: "SdkDefault",
      };
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx);
      expect(got).toEqual(expected);
    });

    it("should error if we expect a number and got another type", async () => {
      const flagKey = "bool_targeting_match";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "TYPE_MISMATCH",
        errorMessage:
          "Flag bool_targeting_match had unexpected type, expected number.",
        value: 123.45,
        flagMetadata: {},
      };
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx);
      expect(got).toEqual(expected);
    });

    it("should error if flag does not exists", async () => {
      const flagKey = "does_not_exists";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "FLAG_NOT_FOUND",
        errorMessage: "Flag with key 'does_not_exists' not found",
        value: 123.45,
        flagMetadata: {},
      };
      const got = await goffClient.getNumberDetails(flagKey, 123.45, userCtx);
      expect(got).toEqual(expected);
    });
  });
  describe("object", () => {
    it("should resolve a valid object flag with TARGETING_MATCH reason", async () => {
      const flagKey = "object_key";
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
        value: {
          test4: 1,
          test2: false,
          test3: 123.3,
          test: "test1",
        },
        variant: "True",
      };
      const got = await goffClient.getObjectDetails(
        flagKey,
        { test: 1234 },
        userCtx
      );
      expect(got).toEqual(expected);
    });

    it("should use object default value if the flag is disabled", async () => {
      const flagKey = "disabled_int";
      const expected = {
        flagKey: flagKey,
        reason: "DISABLED",
        value: { test: 1234 },
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
        variant: "SdkDefault",
      };
      const got = await goffClient.getObjectDetails(
        flagKey,
        { test: 1234 },
        userCtx
      );
      expect(got).toEqual(expected);
    });

    it("should error if we expect an object and got another type", async () => {
      const flagKey = "bool_targeting_match";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "TYPE_MISMATCH",
        errorMessage:
          "Flag bool_targeting_match had unexpected type, expected object.",
        value: { test: 1234 },
        flagMetadata: {},
      };
      const got = await goffClient.getObjectDetails(
        flagKey,
        { test: 1234 },
        userCtx
      );
      expect(got).toEqual(expected);
    });

    it("should error if flag does not exists", async () => {
      const flagKey = "does_not_exists";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        errorCode: "FLAG_NOT_FOUND",
        errorMessage: "Flag with key 'does_not_exists' not found",
        value: { test: 1234 },
        flagMetadata: {},
      };
      const got = await goffClient.getObjectDetails(
        flagKey,
        { test: 1234 },
        userCtx
      );
      expect(got).toEqual(expected);
    });
  });
  describe("authenticated relay proxy", () => {
    beforeEach(async () => {
      // Clean up any existing provider before creating a new one
      await OpenFeature.close();
    });

    afterEach(async () => {
      // Ensure cleanup after each test in this describe block
      await OpenFeature.close();
    });

    it("should resolve a valid flag with an apiKey", async () => {
      const goFeatureFlagProvider = new GoFeatureFlagProvider({
        endpoint: "http://localhost:1032/",
        apiKey: "authorized_token",
      });
      goffClient = OpenFeature.getClient("my-app");
      await OpenFeature.setProviderAndWait("my-app", goFeatureFlagProvider);

      const flagKey = "bool_targeting_match";
      const expected = {
        flagKey: flagKey,
        reason: "TARGETING_MATCH",
        value: true,
        variant: "True",
        flagMetadata: {
          description: "this is a test",
          pr_link: "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
      };
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx);
      expect(got).toEqual(expected);
    });

    it("should resolve a default value with an invalid apiKey", async () => {
      await OpenFeature.close();
      const goFeatureFlagProvider = new GoFeatureFlagProvider({
        endpoint: "http://localhost:1032/",
        apiKey: "invalid-api-key",
      });
      goffClient = OpenFeature.getClient("my-app");
      await expect(
        OpenFeature.setProviderAndWait("my-app", goFeatureFlagProvider)
      ).rejects.toThrow(UnauthorizedError);

      const flagKey = "bool_targeting_match";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "GENERAL",
        flagMetadata: {},
        errorMessage:
          "Provider is not initialized, impossible to retrieve configuration",
      };
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx);
      expect(got).toEqual(expected);
    });

    it("should resolve a default value with an empty apiKey", async () => {
      await OpenFeature.close();
      const goFeatureFlagProvider = new GoFeatureFlagProvider({
        endpoint: "http://localhost:1032/",
        apiKey: "",
      });
      goffClient = OpenFeature.getClient(
        "should-resolve-a-default-value-with-an-empty-apiKey"
      );

      await expect(
        OpenFeature.setProviderAndWait(
          "should-resolve-a-default-value-with-an-empty-apiKey",
          goFeatureFlagProvider
        )
      ).rejects.toThrow(UnauthorizedError);

      const flagKey = "bool_targeting_match";
      const expected = {
        flagKey: flagKey,
        reason: "ERROR",
        value: false,
        errorCode: "GENERAL",
        flagMetadata: {},
        errorMessage:
          "Provider is not initialized, impossible to retrieve configuration",
      };
      const got = await goffClient.getBooleanDetails(flagKey, false, userCtx);
      expect(got).toEqual(expected);
    });
  });
});
