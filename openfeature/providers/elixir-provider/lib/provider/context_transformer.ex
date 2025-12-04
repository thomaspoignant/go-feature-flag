defmodule ElixirProvider.ContextTransformer do
  @moduledoc """
  Converts an OpenFeature EvaluationContext into a GO Feature Flag context.
  """
  alias ElixirProvider.GofEvaluationContext
  alias OpenFeature.Types

  @doc """
  Extracts other key value pairs after the targeting key
  """
  def get_any_value(map) when is_map(map) do
    map
    |> Enum.reject(fn {key, _value} -> key === :targetingKey end)
    |> Enum.into(%{})
  end

  @doc """
  Converts an EvaluationContext map into a ElixirProvider.GofEvaluationContext struct.
  Returns `{:ok, context}` on success, or `{:error, reason}` on failure.
  """
  @spec transform_context(Types.context()) ::
          {:ok, GofEvaluationContext.t()} | {:error, String.t()}
  def transform_context(ctx) do
    case Map.fetch(ctx, :targetingKey) do
      {:ok, value} ->
        {:ok,
         %GofEvaluationContext{
           key: value,
           custom: get_any_value(ctx)
         }}

      :error ->
        {:error, "targeting key not found"}
    end
  end
end
