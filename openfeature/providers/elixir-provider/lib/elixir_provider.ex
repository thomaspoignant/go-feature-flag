defmodule ElixirProvider do
  alias OpenFeature.EvaluationDetails
  alias ElixirProvider.ResponseFlagEvaluation
  alias ElixirProvider.GoFeatureFlagMetadata
  alias ElixirProvider.ContextTransformer
  alias ElixirProvider.RequestFlagEvaluation
  alias ElixirProvider.GoFeatureFlagOptions
  alias ElixirProvider.Types
  alias ElixirProvider.CacheController
  alias ElixirProvider.GoFWebSocketClient
  alias ElixirProvider.HttpClient

  @moduledoc """
  The provider for GO Feature Flag, managing HTTP requests, caching, and flag evaluation.
  """

  defstruct [
    :options,
    :_http_client,
    _data_collector_hook: nil,
    _ws: nil,
  ]

  @type t :: %__MODULE__{
    options: GoFeatureFlagOptions.t(),
    _http_client: HttpClient.t(),
    _data_collector_hook: any(),
    _ws: GoFWebSocketClient.t(),
  }


end
