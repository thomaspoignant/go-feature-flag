defmodule ElixirProviderTest do
  @moduledoc """
  Test file
  """
  use ExUnit.Case, async: true
  require Logger

  doctest ElixirProvider
  alias OpenFeature
  alias OpenFeature.Client

  @endpoint "http://localhost:1031"

  @default_evaluation_ctx %{
    targetingKey: "d45e303a-38c2-11ed-a261-0242ac120002",
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
    _ = start_supervised!(ElixirProvider.ServerSupervisor)

    provider = %ElixirProvider.Provider{
      options: %ElixirProvider.GoFeatureFlagOptions{
        endpoint: @endpoint,
        data_flush_interval: 100,
        disable_cache_invalidation: true
      }
    }

    OpenFeature.set_provider(provider)
    client = OpenFeature.get_client()
    {:ok, client: client}
  end

  ## TEST CONTEXT TRANSFORMER

  test "should use the targetingKey as user key" do
    got =
      ElixirProvider.ContextTransformer.transform_context(%{
        targetingKey: "user-key"
      })

    want(
      = /
        {:ok,
         %ElixirProvider.GofEvaluationContext{
           key: "user-key",
           custom: %{}
         }}
    )

    assert got == want
  end

  test "should specify the anonymous field base on the attributes" do
    got =
      ElixirProvider.ContextTransformer.transform_context(%{
        targetingKey: "user-key",
        anonymous: true
      })

    want =
      {:ok,
       %ElixirProvider.GofEvaluationContext{
         key: "user-key",
         custom: %{
           anonymous: true
         }
       }}

    assert got == want
  end

  test "should fail if no targeting field is provided" do
    got =
      ElixirProvider.ContextTransformer.transform_context(%{
        anonymous: true,
        firstname: "John",
        lastname: "Doe",
        email: "john.doe@gofeatureflag.org"
      })

    want = {:error, "targeting key not found"}

    assert got == want
  end

  test "should fill custom fields if extra fields are present" do
    got =
      ElixirProvider.ContextTransformer.transform_context(%{
        targetingKey: "user-key",
        anonymous: true,
        firstname: "John",
        lastname: "Doe",
        email: "john.doe@gofeatureflag.org"
      })

    want =
      {:ok,
       %ElixirProvider.GofEvaluationContext{
         key: "user-key",
         custom: %{
           firstname: "John",
           lastname: "Doe",
           email: "john.doe@gofeatureflag.org",
           anonymous: true
         }
       }}

    assert got == want
  end

  ## PROVIDER TESTS

  test "should provide an error if flag does not exist", %{client: client} do
    flag_key = "flag_not_found"
    default = false
    ctx = @default_evaluation_ctx

    ElixirProvider.HttpClientMock
    |> expect(:post, fn _client, path, _data ->
      if path == "/v1/feature/#{flag_key}/eval" do
        {:error, {:http_error, 404, "Not Found"}}
      else
        {:error, {:unexpected_path, path}}
      end
    end)

    # Make the client call
    response = Client.get_boolean_details(client, flag_key, default, context: ctx)

    # # Define the expected response structure
    # expected_response = %{
    #   error_code: :provider_not_ready,
    #   error_message:
    #     "impossible to call go-feature-flag relay proxy on #{@endpoint}#{path}: Error: Request failed with status code 404",
    #   key: flag_key,
    #   reason: :error,
    #   value: false,
    #   flag_metadata: %{}
    # }

    # # Assert the response matches the expected structure
    assert response == "?"
  end

  # test "should provide an error if flag does not exist", %{client: client} do
  #   flag_key = "flag_not_found"
  #   default = false
  #   ctx = @default_evaluation_ctx
  #   path = "/v1/feature/#{flag_key}/eval"

  #   # Mock the Mint.HTTP.request/5 function
  #   Mimic.expect(Mint.HTTP, :request, fn _conn, "POST", url, headers, body ->
  #     assert url == "#{@endpoint}#{path}"
  #     assert headers == [{"content-type", "application/json"}]
  #     assert body == Jason.encode!(%{context: ctx, default: default})
  #     {:ok, :mocked_conn, :mocked_request_ref}
  #   end)

  #   # Mock the Mint.HTTP.stream/2 function to simulate a 404 error response
  #   Mimic.expect(Mint.HTTP, :stream, fn _conn, _message ->
  #     {:ok, :mocked_conn,
  #      [
  #        {:status, :mocked_request_ref, 404},
  #        {:headers, :mocked_request_ref, []},
  #        {:data, :mocked_request_ref, ~s<{"error":"flag_not_found"}>},
  #        {:done, :mocked_request_ref}
  #      ]}
  #   end)

  #   # Call the function being tested
  #   response = Client.get_boolean_details(client, flag_key, default, context: ctx)

  #   # Define the expected response
  #   # expected_response = %{
  #   #   error_code: :provider_not_ready,
  #   #   error_message:
  #   #     "impossible to call go-feature-flag relay proxy on #{endpoint}#{path}: Error: Request failed with status code 404",
  #   #   key: flag_key,
  #   #   reason: :error,
  #   #   value: false,
  #   #   flag_metadata: %{}
  #   # }

  #   # Assert the response matches the expected response
  #   # assert response == "?"
  # end

  test "post/3 sends a POST request and processes the response" do
    # Mock the Mint.HTTP.request/5 function
    Mimic.expect(Mint.HTTP, :request, fn _conn, "POST", url, headers, body ->
      assert url == "https://api.example.com/v1/test/path"
      assert headers == [{"content-type", "application/json"}]
      assert body == ~s<{"key":"value"}>
      {:ok, :mocked_conn, :mocked_request_ref}
    end)

    # Mock the Mint.HTTP.stream/2 function to simulate a 200 OK response
    Mimic.expect(Mint.HTTP, :stream, fn _conn, _message ->
      {:ok, :mocked_conn,
       [
         {:status, :mocked_request_ref, 200},
         {:headers, :mocked_request_ref, []},
         {:data, :mocked_request_ref, ~s<{"message":"success"}>},
         {:done, :mocked_request_ref}
       ]}
    end)

    # Prepare the connection struct
    client = %ElixirProvider.HttpClient{
      conn: :mocked_conn,
      endpoint: "https://api.example.com",
      headers: [{"content-type", "application/json"}]
    }

    # Call the post/3 function
    response = ElixirProvider.HttpClient.post(client, "/v1/test/path", %{"key" => "value"})

    # Assert the decoded response
    assert {:ok, %{"message" => "success"}} == response
  end
end
