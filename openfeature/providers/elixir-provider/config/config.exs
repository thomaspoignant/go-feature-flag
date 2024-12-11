import Config

config :elixir_provider,
  max_wait_time: 5000,
  hackney_options: [timeout: :infinity, recv_timeout: :infinity]

import_config "#{config_env()}.exs"
