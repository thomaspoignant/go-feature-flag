using System.Runtime.InteropServices.JavaScript;
using OpenFeature.Model;
using OpenFeature;
using OpenFeature.Constant;
using OpenFeature.Contrib.Providers.GOFeatureFlag;
using FluentAssertions;

namespace ProviderTests;


public class ProviderTests
{
    private GoFeatureFlagProvider goFeatureFlagProvider;
    private FeatureClient client;
    private EvaluationContext defaultEvaluationContext;
    [SetUp]
    public void Setup()
    {
        goFeatureFlagProvider = new GoFeatureFlagProvider(new GoFeatureFlagProviderOptions
        {
            Endpoint = "http://localhost:1031/",
            Timeout = new TimeSpan(1000 * TimeSpan.TicksPerMillisecond)
        });
        Api.Instance.SetProvider(goFeatureFlagProvider);
        client = Api.Instance.GetClient("my-app");
        
        defaultEvaluationContext = EvaluationContext.Builder()
            .Set("targetingKey", "d45e303a-38c2-11ed-a261-0242ac120002") // user unique identifier (mandatory)
            .Set("email", "john.doe@gofeatureflag.org")
            .Set("firstname", "john")
            .Set("lastname", "doe")
            .Set("anonymous", false)
            .Set("professional", true)
            .Set("rate", 3.14)
            .Set("age", 30)
            .Set("admin", true)
            .Set("labels", new Value(new List<Value> { new("pro"), new("beta") }))
            .Set("company_info", Structure.Builder().Set("name","my_company").Set("size",120).Build())
            .Build();
    }

    [Test]
    public async Task ShouldResolveAValidBooleanFlagWithTargetingMatchReason()
    {
        var flagKey = "bool_targeting_match";
        var want = new FlagEvaluationDetails<bool>(
            flagKey, 
            true, 
            ErrorType.None, 
            Reason.TargetingMatch,
            "True",
            null);
        var got = await client.GetBooleanDetails(flagKey, false, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);

    }
    [Test]
    public async Task ShouldUseBooleanDefaultValueIfTheFlagIsDisabled()
    {
        var flagKey = "disabled_bool";
        var want = new FlagEvaluationDetails<bool>(
            flagKey, 
            false, 
            ErrorType.None, 
            Reason.Disabled,
            null,
            null); 
        var got = await client.GetBooleanDetails(flagKey, false, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfWeExpectABooleanAndGotAnotherType()
    {
        var flagKey = "string_key";
        var want = new FlagEvaluationDetails<bool>(
            flagKey, 
            false, 
            ErrorType.TypeMismatch, 
            Reason.Error,
            "",
            "flag value string_key had unexpected type"); 
        var got = await client.GetBooleanDetails(flagKey, false, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfFlagDoesNotExistsBool()
    {
        var flagKey = "does_not_exists";
        var want = new FlagEvaluationDetails<bool>(
            flagKey, 
            false, 
            ErrorType.FlagNotFound, 
            Reason.Error,
            "",
            $"flag {flagKey} was not found in your configuration"); 
        var got = await client.GetBooleanDetails(flagKey, false, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldResolveAValidStringFlagWithTargetingMatchReason()
    {
        var flagKey = "string_key";
        var want = new FlagEvaluationDetails<string>(
            flagKey, 
            "CC0000", 
            ErrorType.None, 
            Reason.TargetingMatch,
            "True",
            null); 
        var got = await client.GetStringDetails(flagKey, "default", defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldUseStringDefaultValueIfTheFlagIsDisabled()
    {
        var flagKey = "disabled_string";
        var want = new FlagEvaluationDetails<string>(
            flagKey, 
            "default", 
            ErrorType.None, 
            Reason.Disabled,
            null,
            null); 
        var got = await client.GetStringDetails(flagKey, "default", defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfWeExpectAStringAndGotAnotherType()
    {
        var flagKey = "bool_targeting_match";
        var want = new FlagEvaluationDetails<string>(
            flagKey, 
            "default", 
            ErrorType.TypeMismatch, 
            Reason.Error,
            "",
            "flag value bool_targeting_match had unexpected type"); 
        var got = await client.GetStringDetails(flagKey, "default", defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfFlagDoesNotExistsString()
    {
        var flagKey = "does_not_exists";
        var want = new FlagEvaluationDetails<string>(
            flagKey, 
            "default", 
            ErrorType.FlagNotFound, 
            Reason.Error,
            "",
            $"flag {flagKey} was not found in your configuration"); 
        var got = await client.GetStringDetails(flagKey, "default", defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldResolveAValidDoubleFlagWithTargetingMatchReason()
    {
        var flagKey = "double_key";
        var want = new FlagEvaluationDetails<double>(
            flagKey, 
            100.25,
            ErrorType.None, 
            Reason.TargetingMatch,
            "True",
            null); 
        var got = await client.GetDoubleDetails(flagKey,  123.45, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldUseDoubleDefaultValueIfTheFlagIsDisabled()
    {
        var flagKey = "disabled_float";
        var want = new FlagEvaluationDetails<double>(
            flagKey, 
            123.45,
            ErrorType.None, 
            Reason.Disabled,
            null,
            null); 
        var got = await client.GetDoubleDetails(flagKey,  123.45, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfWeExpectADoubleAndGotAnotherType()
    {
        var flagKey = "bool_targeting_match";
        var want = new FlagEvaluationDetails<double>(
            flagKey, 
            123.45,
            ErrorType.TypeMismatch, 
            Reason.Error,
            "",
            "flag value bool_targeting_match had unexpected type"); 
        var got = await client.GetDoubleDetails(flagKey,  123.45, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfFlagDoesNotExistsDouble()
    {
        var flagKey = "does_not_exists";
        var want = new FlagEvaluationDetails<double>(
            flagKey, 
            123.45,
            ErrorType.FlagNotFound, 
            Reason.Error,
            "",
            $"flag {flagKey} was not found in your configuration"); 
        var got = await client.GetDoubleDetails(flagKey, 123.45, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldResolveAValidIntFlagWithTargetingMatchReason()
    {
        var flagKey = "integer_key";
        var want = new FlagEvaluationDetails<int>(
            flagKey, 
            100,
            ErrorType.None, 
            Reason.TargetingMatch,
            "True",
            null); 
        var got = await client.GetIntegerDetails(flagKey,  123, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldUseIntDefaultValueIfTheFlagIsDisabled()
    {
        var flagKey = "disabled_int";
        var want = new FlagEvaluationDetails<int>(
            flagKey, 
            123,
            ErrorType.None, 
            Reason.Disabled,
            null,
            null); 
        var got = await client.GetIntegerDetails(flagKey,  123, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfWeExpectAIntAndGotAnotherType()
    {
        var flagKey = "bool_targeting_match";
        var want = new FlagEvaluationDetails<int>(
            flagKey, 
            123,
            ErrorType.TypeMismatch, 
            Reason.Error,
            "",
            "flag value bool_targeting_match had unexpected type"); 
        var got = await client.GetIntegerDetails(flagKey,  123, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfFlagDoesNotExistsInt()
    {
        var flagKey = "does_not_exists";
        var want = new FlagEvaluationDetails<int>(
            flagKey, 
            123,
            ErrorType.FlagNotFound, 
            Reason.Error,
            "",
            $"flag {flagKey} was not found in your configuration"); 
        var got = await client.GetIntegerDetails(flagKey,  123, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldResolveAValidObjectFlagWithTargetingMatchReason()
    {
        Structure expectedStructure = new Structure(
        new Dictionary<string, Value>()
            {
                { "test4", new Value(1) },
                { "test2", new Value(false) },
                {"test3", new Value(123.3)},
                {"test", new Value("test1")},
            }
        );
        var flagKey = "object_key";
        var want = new FlagEvaluationDetails<Value>(
            flagKey, 
            new Value(expectedStructure),
            ErrorType.None, 
            Reason.TargetingMatch,
            "True",
            null); 
        var got = await client.GetObjectDetails(flagKey,  new Value(123), defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldUseObjectDefaultValueIfTheFlagIsDisabled()
    {
        var flagKey = "disabled_int";
        var want = new FlagEvaluationDetails<Value>(
            flagKey, 
            new Value(123),
            ErrorType.None, 
            Reason.Disabled,
            null,
            null); 
        var got = await client.GetObjectDetails(flagKey,  new Value(123), defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task ShouldErrorIfFlagDoesNotExistsObject()
    {
        var flagKey = "does_not_exists";
        var want = new FlagEvaluationDetails<Value>(
            flagKey, 
            new Value(123),
            ErrorType.FlagNotFound, 
            Reason.Error,
            "",
            $"flag {flagKey} was not found in your configuration"); 
        var got = await client.GetObjectDetails(flagKey,  new Value(123), defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task Should_resolve_a_valid_flag_with_an_apiKey()
    {
        GoFeatureFlagProvider authenticatedProvider = new GoFeatureFlagProvider(new GoFeatureFlagProviderOptions
        {
            Endpoint = "http://localhost:1032/",
            Timeout = new TimeSpan(1000 * TimeSpan.TicksPerMillisecond),
            ApiKey = "authorized_token"
        });
        Api.Instance.SetProvider(authenticatedProvider);
        FeatureClient authenticatedClient = Api.Instance.GetClient("my-app");
        var flagKey = "bool_targeting_match";
        var want = new FlagEvaluationDetails<bool>(
            flagKey, 
            true, 
            ErrorType.None, 
            Reason.TargetingMatch,
            "True",
            null);
        var got = await authenticatedClient.GetBooleanDetails(flagKey, false, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    
    [Test]
    public async Task Should_resolve_a_default_value_with_an_invalid_apiKey()
    {
        GoFeatureFlagProvider authenticatedProvider = new GoFeatureFlagProvider(new GoFeatureFlagProviderOptions
        {
            Endpoint = "http://localhost:1032/",
            Timeout = new TimeSpan(1000 * TimeSpan.TicksPerMillisecond),
            ApiKey = "invalid_api_key"
        });
        Api.Instance.SetProvider(authenticatedProvider);
        FeatureClient authenticatedClient = Api.Instance.GetClient("my-app");
        var flagKey = "bool_targeting_match";
        var want = new FlagEvaluationDetails<bool>(
            flagKey, 
            false, 
            ErrorType.General, 
            Reason.Error,
            "True",
            "invalid api key, impossible to authenticate the provider");
        var got = await authenticatedClient.GetBooleanDetails(flagKey, false, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
    [Test]
    public async Task Should_resolve_a_default_value_with_an_empty_apiKey()
    {
        GoFeatureFlagProvider authenticatedProvider = new GoFeatureFlagProvider(new GoFeatureFlagProviderOptions
        {
            Endpoint = "http://localhost:1032/",
            Timeout = new TimeSpan(1000 * TimeSpan.TicksPerMillisecond),
            ApiKey = ""
        });
        Api.Instance.SetProvider(authenticatedProvider);
        FeatureClient authenticatedClient = Api.Instance.GetClient("my-app");
        var flagKey = "bool_targeting_match";
        var want = new FlagEvaluationDetails<bool>(
            flagKey, 
            false, 
            ErrorType.General, 
            Reason.Error,
            "True",
            "invalid api key, impossible to authenticate the provider");
        var got = await authenticatedClient.GetBooleanDetails(flagKey, false, defaultEvaluationContext);
        want.Should().BeEquivalentTo(got);
    }
}