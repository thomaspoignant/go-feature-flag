defmodule ElixirProvider.GoFeatureFlagMetadata do
  @moduledoc """
  Metadata for the Go Feature Flag.
  """

  defstruct [name: "Go Feature Flag"]

  @type t :: %__MODULE__{
          name: String.t()
        }
end
