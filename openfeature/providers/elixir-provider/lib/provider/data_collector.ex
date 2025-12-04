defmodule ElixirProvider.RequestDataCollector do
  @moduledoc """
  Represents the data collected in a request, including meta information and events.
  """
  alias ElixirProvider.FeatureEvent

  defstruct [:meta, events: []]

  @type t :: %__MODULE__{
          meta: %{optional(String.t()) => String.t()},
          events: [FeatureEvent.t()]
        }
end
