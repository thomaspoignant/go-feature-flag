defmodule ElixirProviderTest do
  @moduledoc """
  Test file
  """
  use ExUnit.Case
  doctest ElixirProvider

  ## TEST CONTEXT TRANSFORMER

  test "should use the targetingKey as user key" do
    got =
      ElixirProvider.ContextTransformer.transform_context(%{
        targetingKey: "user-key"
      })

    want =
      {:ok,
       %ElixirProvider.GofEvaluationContext{
         key: "user-key",
         custom: %{}
       }}

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

  ### PROVIDER TESTS
end
