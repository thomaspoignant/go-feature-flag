defmodule ElixirProvider.CacheController do
  @moduledoc """
  Controller for caching flag evaluations to avoid redundant API calls.
  """

  use GenServer
  @flag_table :flag_cache

  @spec start_link(Keyword.t()) :: GenServer.on_start()
  def start_link(opts) do
    name = Keyword.get(opts, :name, __MODULE__)
    GenServer.start_link(__MODULE__, :ok, name: name)
  end

  def get(flag_key, evaluation_hash) do
    cache_key = build_cache_key(flag_key, evaluation_hash)
    case :ets.lookup(@flag_table, cache_key) do
      [{^cache_key, cached_value}] -> {:ok, cached_value}
      [] -> :miss
    end
  end

  def set(flag_key, evaluation_hash, value) do
    cache_key = build_cache_key(flag_key, evaluation_hash)
    :ets.insert(@flag_table, {cache_key, value})
    :ok
  end

  def clear do
    :ets.delete_all_objects(@flag_table)
    :ets.insert(@flag_table, {:context, %{}})
    :ok
  end

  defp build_cache_key(flag_key, evaluation_hash) do
    "#{flag_key}-#{evaluation_hash}"
  end

  @impl true
  def init(:ok) do
    :ets.new(@flag_table, [:named_table, :set, :public])
    :ets.insert(@flag_table, {:context, %{}})
    {:ok, nil, :hibernate}
  end
end
