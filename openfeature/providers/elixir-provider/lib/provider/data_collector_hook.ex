defmodule ElixirProvider.DataCollectorHook do
  @moduledoc """
  Data collector hook
  """

  use GenServer
  require Logger

  alias ElixirProvider.HttpClient
  alias ElixirProvider.{FeatureEvent, RequestDataCollector}

  @default_targeting_key "undefined-targetingKey"

  defstruct [
    :http_client,
    :data_collector_endpoint,
    :disable_data_collection,
    data_flush_interval: 60_000,
    event_queue: []
  ]

  @type t :: %__MODULE__{
          http_client: HttpClient.t(),
          data_collector_endpoint: String.t(),
          disable_data_collection: boolean(),
          data_flush_interval: non_neg_integer(),
          event_queue: list(FeatureEvent.t())
        }

  # Starts the GenServer and initializes with options
  def start_link do
    GenServer.start_link(__MODULE__, [], name: __MODULE__)
  end

  def stop(state) do
    GenServer.stop(__MODULE__)
    collect_data(state.data_flush_interval)

    %__MODULE__{
      http_client: state.http_client,
      data_collector_endpoint: state.data_collector_endpoint,
      disable_data_collection: state.disable_data_collection,
      data_flush_interval: state.data_flush_interval,
      event_queue: []
    }
  end

  @impl true
  def init([]) do
    {:ok, %__MODULE__{}}
  end

  # Initializes the state with the provided options
  def start(options, http_client) do
    state = %__MODULE__{
      http_client: http_client,
      data_collector_endpoint: options.endpoint,
      disable_data_collection: options.disable_data_collection || false,
      data_flush_interval: options.data_flush_interval || 60_000,
      event_queue: []
    }

    schedule_collect_data(state.data_flush_interval)
    {:ok, state}
  end

  # Schedule periodic data collection based on the interval
  defp schedule_collect_data(interval) do
    Process.send_after(self(), :collect_data, interval)
  end

  ### Hook Implementations
  def after_hook(hook, hook_context, flag_evaluation_details, _hints) do
    if hook.disable_data_collection or flag_evaluation_details.reason != :CACHED do
      :ok
    else
      feature_event = %FeatureEvent{
        context_kind:
          if(Map.get(hook_context.context, "anonymous"), do: "anonymousUser", else: "user"),
        creation_date: DateTime.utc_now() |> DateTime.to_unix(:millisecond),
        default: false,
        key: hook_context.flag_key,
        value: flag_evaluation_details.value,
        variation: flag_evaluation_details.variant || "SdkDefault",
        user_key:
          Map.get(hook_context.evaluation_context, "targeting_key") || @default_targeting_key
      }

      GenServer.cast(__MODULE__, {:add_event, feature_event})
    end
  end

  def error(hook, hook_context, _hints) do
    if hook.disable_data_collection do
      :ok
    else
      feature_event = %FeatureEvent{
        context_kind:
          if(Map.get(hook_context.context, "anonymous"), do: "anonymousUser", else: "user"),
        creation_date: DateTime.utc_now() |> DateTime.to_unix(:millisecond),
        default: true,
        key: hook_context.flag_key,
        value: Map.get(hook_context.context, "default_value"),
        variation: "SdkDefault",
        user_key: Map.get(hook_context.context, "targeting_key") || @default_targeting_key
      }

      GenServer.call(__MODULE__, {:add_event, feature_event})
    end
  end

  ### GenServer Callbacks
  @impl true
  def handle_call({:add_event, feature_event}, _from, state) do
    {:reply, :ok, %{state | event_queue: [feature_event | state.event_queue]}}
  end

  # Handle the periodic flush
  @impl true
  def handle_info(:collect_data, state) do
    case collect_data(state) do
      :ok -> Logger.info("Data collected and sent successfully.")
      {:error, reason} -> Logger.error("Failed to send data: #{inspect(reason)}")
    end

    schedule_collect_data(state.data_flush_interval)
    {:noreply, %{state | event_queue: []}}
  end

  defp collect_data(%__MODULE__{
         event_queue: event_queue,
         http_client: http_client,
         data_collector_endpoint: endpoint
       }) do
    if Enum.empty?(event_queue) do
      :ok
    else
      body = %RequestDataCollector{
        meta: %{"provider" => "open-feature-elixir-sdk"},
        events: event_queue
      }

      case HttpClient.post(http_client, endpoint, body) do
        {:ok, response} ->
          Logger.info("Data sent successfully: #{inspect(response)}")
          :ok

        {:error, reason} ->
          Logger.error("Error sending data: #{inspect(reason)}")
          {:error, reason}
      end
    end
  end
end
