defmodule ElixirProvider.ResponseFlagEvaluation do
  @moduledoc """
  Represents the evaluation response of a feature flag.
  """
  alias ElixirProvider.Types

  @enforce_keys [:value, :failed, :reason]
  defstruct [:value, error_code: nil,
            failed: false,
            reason: "",
            track_events: nil,
            variation_type: nil,
            version: nil,
            metadata: nil,
            cacheable: nil
          ]

  @type t :: %__MODULE__{
          error_code: String.t() | nil,
          failed: boolean(),
          reason: String.t(),
          track_events: boolean() | nil,
          value: Types.json_type(),
          variation_type: String.t() | nil,
          version: String.t() | nil,
          metadata: map() | nil,
          cacheable: boolean() | nil
        }
end
