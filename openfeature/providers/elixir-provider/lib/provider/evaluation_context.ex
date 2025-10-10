defmodule ElixirProvider.GofEvaluationContext do
  @moduledoc """
  GoFeatureFlagEvaluationContext is an object representing a user context for evaluation.
  """
  alias Jason
  @derive Jason.Encoder
  defstruct key: "", custom: %{}

  @type t :: %__MODULE__{
          key: String.t(),
          custom: map() | nil
        }

  @doc """
  Generates an MD5 hash based on the `key` and `custom` fields.
  """
  def hash(%__MODULE__{key: key, custom: custom}) do
    data = %{"key" => key, "custom" => custom}
    encoded = Jason.encode!(data, pretty: true)
    :crypto.hash(:md5, encoded) |> Base.encode16(case: :lower)
  end
end
