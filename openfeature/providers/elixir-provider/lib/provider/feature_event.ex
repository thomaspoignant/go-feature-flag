defmodule ElixirProvider.FeatureEvent do
  @moduledoc """
  Represents a feature event with details about the feature flag evaluation.
  """
  @enforce_keys [:kind, :context_kind, :user_key, :creation_date, :key, :variation]
  defstruct [kind: "feature",
            context_kind: "",
            user_key: "",
            creation_date: 0,
            key: "",
            variation: "",
            value: nil,
            default: false,
            source: "PROVIDER_CACHE"]

  @type t :: %__MODULE__{
          kind: String.t(),
          context_kind: String.t(),
          user_key: String.t(),
          creation_date: integer(),
          key: String.t(),
          variation: String.t(),
          value: any(),
          default: boolean(),
          source: String.t()
        }
end
