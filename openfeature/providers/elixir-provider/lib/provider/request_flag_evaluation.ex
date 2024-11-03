defmodule ElixirProvider.RequestFlagEvaluation do
  @moduledoc """
  RequestFlagEvaluation is an object representing a user context for evaluation.
  """
  alias ElixirProvider.GofEvaluationContext

  @enforce_keys [:user]
  defstruct [:default_value, :user]

  @type t :: %__MODULE__{
          user: GofEvaluationContext.t(),
          default_value: any()
        }
end
