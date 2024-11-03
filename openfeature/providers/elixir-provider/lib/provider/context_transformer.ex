defmodule ElixirProvider.ContextTransformer do
  @moduledoc """
  Converts an OpenFeature EvaluationContext into a GO Feature Flag context.
  """
  alias ElixirProvider.EvaluationContext
  alias OpenFeature.Types

  @doc """
  Finds any key-value pair with a non-nil value.
  """
  def get_any_value(map) when is_map(map) do
    case Enum.find(map, fn {_key, value} -> value != nil end) do
      {key, value} -> {:ok, {key, value}}
      nil -> {:error, "No keys found with a value"}
    end
  end

  @doc """
  Converts an EvaluationContext map into a ElixirProvider.EvaluationContext struct.
  Returns `{:ok, context}` on success, or `{:error, reason}` on failure.
  """
  @spec transform_context(Types.context()) :: {:ok, EvaluationContext.t()} | {:error, String.t()}
  def transform_context(ctx) do
    case get_any_value(ctx) do
      {:ok, {key, value}} ->
        {:ok, %EvaluationContext{
          key: key,
          custom: value
        }}
      {:error, reason} ->
        {:error, reason}
    end
  end
end
