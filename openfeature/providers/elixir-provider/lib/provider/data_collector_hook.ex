defmodule ElixirProvider.DataCollectorHook do

  require Logger

  alias ElixirProvider.HttpClient

  @moduledoc """
  Handles data collection by buffering events and sending them to the relay proxy at intervals.
  """

  defstruct [
    :goff_api_controller,
    :data_flush_interval,
    :data_collector_metadata,
    :collect_uncached_evaluation,
    event_queue: []
  ]

  @type t :: %__MODULE__{
          goff_api_controller: HttpClient.t(),
          data_flush_interval: non_neg_integer(),
          data_collector_metadata: map(),
          collect_uncached_evaluation: boolean(),
          event_queue: list()
        }

  end
