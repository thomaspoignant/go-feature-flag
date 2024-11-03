defmodule ElixirProvider do
  use WebSockex

  alias OpenFeature.EvaluationDetails
  alias ElixirProvider.ResponseFlagEvaluation
  alias ElixirProvider.GoFeatureFlagMetadata
  alias ElixirProvider.ContextTransformer
  alias ElixirProvider.RequestFlagEvaluation
  alias ElixirProvider.GoFeatureFlagOptions
  alias ElixirProvider.Types
  alias ElixirProvider.CacheController

  @moduledoc """
  The provider for GO Feature Flag, managing HTTP requests, caching, and flag evaluation.
  """

  defstruct [
    :options,
    :_http_client,
    _cache_controller: nil,
    _data_collector_hook: nil,
    _ws: nil,
  ]

  @type t :: %__MODULE__{
    options: GoFeatureFlagOptions,
    _cache_controller: CacheController,
    _http_client: any(),
    _data_collector_hook: any(),
    _ws: WebSockex,
  }


end
