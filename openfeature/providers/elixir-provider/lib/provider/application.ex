defmodule OpenFeature.Application do
  @moduledoc false

  use Application

  @impl true
  def start(_type, _args) do
    children = [
      OpenFeature.Store,
      OpenFeature.EventEmitter
    ]

    opts = [strategy: :one_for_one, name: OpenFeature.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
