defmodule ElixirProvider.Provider do
  @behaviour OpenFeature.Provider

  alias OpenFeature.ResolutionDetails
  alias ElixirProvider.GoFeatureFlagOptions
  alias ElixirProvider.HttpClient
  alias ElixirProvider.DataCollectorHook
  alias ElixirProvider.CacheController
  alias ElixirProvider.ResponseFlagEvaluation
  alias ElixirProvider.GoFWebSocketClient
  alias ElixirProvider.RequestFlagEvaluation
  alias ElixirProvider.ContextTransformer
  alias ElixirProvider.GofEvaluationContext

  @moduledoc """
  The GO Feature Flag provider for OpenFeature, managing HTTP requests, caching, and flag evaluation.
  """

  defstruct [
    :options,
    :http_client,
    :data_collector_hook,
    :ws,
    :domain
  ]

  @type t :: %__MODULE__{
          options: GoFeatureFlagOptions.t(),
          http_client: HttpClient.t(),
          data_collector_hook: DataCollectorHook.t() | nil,
          ws: GoFWebSocketClient.t(),
          domain: String.t()
        }

  @impl true
  def initialize(%__MODULE__{} = provider, domain, _context) do
    {:ok, http_client} = HttpClient.start_http_connection(provider.options)
    {:ok, data_collector_hook} = DataCollectorHook.start(provider.options, http_client)
    {:ok, ws} = GoFWebSocketClient.connect(provider.options.endpoint)

    updated_provider = %__MODULE__{
      provider
      | domain: domain,
        http_client: http_client,
        data_collector_hook: data_collector_hook,
        ws: ws
    }

    {:ok, updated_provider}
  end

  @impl true
  def shutdown(%__MODULE__{ws: ws} = provider) do
    Process.exit(ws, :normal)
    CacheController.clear()
    if(GenServer.whereis(GoFWebSocketClient), do: GoFWebSocketClient.stop())
    if(GenServer.whereis(DataCollectorHook), do: DataCollectorHook.stop(provider.data_collector_hook))
    :ok
  end

  @impl true
  def resolve_boolean_value(provider, key, default, context) do
    generic_resolve(provider, :boolean, key, default, context)
  end

  @impl true
  def resolve_string_value(provider, key, default, context) do
    generic_resolve(provider, :string, key, default, context)
  end

  @impl true
  def resolve_number_value(provider, key, default, context) do
    generic_resolve(provider, :number, key, default, context)
  end

  @impl true
  def resolve_map_value(provider, key, default, context) do
    generic_resolve(provider, :map, key, default, context)
  end

  defp generic_resolve(provider, type, flag_key, default_value, context) do
    {:ok, goff_context} = ContextTransformer.transform_context(context)
    goff_request = %RequestFlagEvaluation{user: goff_context, default_value: default_value}
    eval_context_hash = GofEvaluationContext.hash(goff_context)

    response_body =
      case CacheController.get(flag_key, eval_context_hash) do
        {:ok, cached_response} ->
          cached_response

        :miss ->
          # Fetch from HTTP if cache miss
          case HttpClient.post(provider.http_client, "/v1/feature/#{flag_key}/eval", goff_request) do
            {:ok, response} -> handle_response(flag_key, eval_context_hash, response)
            {:error, reason} -> {:error, {:unexpected_error, reason}}
          end
      end

    handle_flag_resolution(response_body, type, flag_key, default_value)
  end

  defp handle_response(flag_key, eval_context_hash, response) do
    # Build the flag evaluation struct directly from the response map
    flag_eval = ResponseFlagEvaluation.decode(response)

    # Cache the response if it's marked as cacheable
    if flag_eval.cacheable do
      CacheController.set(flag_key, eval_context_hash, response)
    end

    {:ok, flag_eval}
  end

  defp handle_flag_resolution(response, type, flag_key, _default_value) do
    case response do
      {:ok, %ResponseFlagEvaluation{value: value, reason: reason}} ->
        case {type, value} do
          {:boolean, val} when is_boolean(val) ->
            {:ok, %ResolutionDetails{value: val, reason: reason}}

          {:string, val} when is_binary(val) ->
            {:ok, %ResolutionDetails{value: val, reason: reason}}

          {:number, val} when is_number(val) ->
            {:ok, %ResolutionDetails{value: val, reason: reason}}

          {:map, val} when is_map(val) ->
            {:ok, %ResolutionDetails{value: val, reason: reason}}

          _ ->
            {:error, {:variant_not_found, "Expected #{type} but got #{inspect(value)} for flag #{flag_key}"}}
        end

      _ ->
        {:error, {:flag_not_found, "Flag #{flag_key} not found"}}
    end
  end

end
