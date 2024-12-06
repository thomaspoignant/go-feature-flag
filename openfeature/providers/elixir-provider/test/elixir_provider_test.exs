defmodule ElixirProviderTest do
  @moduledoc """
  Test file
  """
  use ExUnit.Case
  doctest ElixirProvider
  alias OpenFeature
  alias OpenFeature.Client

  @endpoint "http://localhost:1031"

  @default_evaluation_ctx %{
    targeting_key: "d45e303a-38c2-11ed-a261-0242ac120002",
    email: "john.doe@gofeatureflag.org",
    firstname: "john",
    lastname: "doe",
    anonymous: false,
    professional: true,
    rate: 3.14,
    age: 30,
    company_info: %{name: "my_company", size: 120},
    labels: ["pro", "beta"]
  }

  setup do
    provider = %ElixirProvider.Provider{
      options: %ElixirProvider.GoFeatureFlagOptions{
        endpoint: @endpoint,
        data_flush_interval: 100,
        disable_cache_invalidation: true
      }
    }

    bypass = Bypass.open()
    OpenFeature.set_provider(provider)
    client = OpenFeature.get_client()
    {:ok, bypass: bypass, client: client}
  end

  ## TEST CONTEXT TRANSFORMER

  # test "should use the targetingKey as user key" do
  #   got =
  #     ElixirProvider.ContextTransformer.transform_context(%{
  #       targetingKey: "user-key"
  #     })

  #   want =/
  #     {:ok,
  #      %ElixirProvider.GofEvaluationContext{
  #        key: "user-key",
  #        custom: %{}
  #      }}

  #   assert got == want
  # end

  # test "should specify the anonymous field base on the attributes" do
  #   got =
  #     ElixirProvider.ContextTransformer.transform_context(%{
  #       targetingKey: "user-key",
  #       anonymous: true
  #     })

  #   want =
  #     {:ok,
  #      %ElixirProvider.GofEvaluationContext{
  #        key: "user-key",
  #        custom: %{
  #          anonymous: true
  #        }
  #      }}

  #   assert got == want
  # end

  # test "should fail if no targeting field is provided" do
  #   got =
  #     ElixirProvider.ContextTransformer.transform_context(%{
  #       anonymous: true,
  #       firstname: "John",
  #       lastname: "Doe",
  #       email: "john.doe@gofeatureflag.org"
  #     })

  #   want = {:error, "targeting key not found"}

  #   assert got == want
  # end

  # test "should fill custom fields if extra fields are present" do
  #   got =
  #     ElixirProvider.ContextTransformer.transform_context(%{
  #       targetingKey: "user-key",
  #       anonymous: true,
  #       firstname: "John",
  #       lastname: "Doe",
  #       email: "john.doe@gofeatureflag.org"
  #     })

  #   want =
  #     {:ok,
  #      %ElixirProvider.GofEvaluationContext{
  #        key: "user-key",
  #        custom: %{
  #          firstname: "John",
  #          lastname: "Doe",
  #          email: "john.doe@gofeatureflag.org",
  #          anonymous: true
  #        }
  #      }}

  #   assert got == want
  # end

  ### PROVIDER TESTS

  test "should provide an error if flag does not exist", %{bypass: bypass, client: client} do
    flag_key = "flag_not_found"
    default = false
    ctx = @default_evaluation_ctx

    # Corrected path (only the path, not the full URL)
    path = "/v1/feature/#{flag_key}/eval"

    # Set up Bypass to handle the POST request
    Bypass.expect_once(bypass, "POST", path, fn conn ->
      Plug.Conn.resp(conn, 404, ~s<{"errors": [{"code": 88, "message": "Rate limit exceeded"}]}>)
    end)

    # Make the client call
    response = Client.get_boolean_details(client, flag_key, default, context: ctx)

    # Define the expected response structure
    expected_response = %{
      error_code: :provider_not_ready,
      error_message:
        "impossible to call go-feature-flag relay proxy on #{@endpoint}#{path}: Error: Request failed with status code 404",
      key: flag_key,
      reason: :error,
      value: false,
      flag_metadata: %{}
    }

    # Assert the response matches the expected structure
    assert response == expected_response
  end
end
