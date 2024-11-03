defmodule ElixirProvider.GoFeatureFlagOptions do
  @moduledoc """
  Configuration options for the Go Feature Flag.
  """

  @enforce_keys [:endpoint]
  defstruct [:endpoint,
            cache_size: 10_000,
            data_flush_interval: 60_000,
            disable_data_collection: false,
            reconnect_interval: 60,
            disable_cache_invalidation: false]

  @type t :: %__MODULE__{
          endpoint: String.t(),
          cache_size: integer() | nil,
          data_flush_interval: integer() | nil,
          disable_data_collection: boolean(),
          reconnect_interval: integer() | nil,
          disable_cache_invalidation: boolean() | nil
        }
end
