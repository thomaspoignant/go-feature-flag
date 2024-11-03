defmodule ElixirProvider.GoFWebSocketClient do
  use GenServer
  require Logger

  alias ElixirProvider.CacheController

  @moduledoc """
  A minimal WebSocket client for listening to configuration changes from the GO Feature Flag relay proxy.
  Clears the cache on receiving change notifications.
  """

  @type t :: Mint.WebSocket.t()

  @websocket_uri "/ws/v1/flag/change"

  # Public API

  # Start the WebSocket client with a URL
  def start_link(url) do
    GenServer.start_link(__MODULE__, url, name: __MODULE__)
  end

  # GenServer Callbacks

  def init(url) do
    state = %{}

    # Connect to the WebSocket server
    case connect(url) do
      {:ok, conn, websocket} ->
        Logger.info("Connected to WebSocket at #{url}")
        {:ok, %{conn: conn, websocket: websocket, url: url}}

      {:error, reason} ->
        Logger.error("Failed to connect to WebSocket: #{inspect(reason)}")
        {:stop, reason, state}
    end
  end

  # Handle incoming messages and check for change notifications
  def handle_info({:websocket, {:text, message}}, state) do
    case Jason.decode(message) do
      {:ok, %{"type" => "change"}} ->
        # Clear the cache when a change message is received
        CacheController.clear()
        Logger.info("Cache cleared due to configuration change notification.")

      _other ->
        Logger.debug("Received non-change message: #{message}")
    end

    {:noreply, state}
  end

  # Handle WebSocket disconnection and attempt reconnection
  def handle_info({:websocket, :closed}, %{url: url} = state) do
    Logger.warning("WebSocket disconnected. Attempting to reconnect...")
    case connect(url) do
      {:ok, conn, websocket} ->
        {:noreply, %{state | conn: conn, websocket: websocket}}

      {:error, reason} ->
        Logger.error("Failed to reconnect: #{inspect(reason)}")
        {:stop, reason, state}
    end
  end

  # Private Helper Functions

  defp connect(url) do
    uri = URI.parse(url)
    http_scheme = if uri.scheme == "ws", do: :http, else: :https
    websocket_scheme = if uri.scheme == "ws", do: :ws, else: :wss

    # Construct the WebSocket path
    path = uri.path <> @websocket_uri

    with {:ok, conn} <- Mint.HTTP.connect(http_scheme, uri.host, uri.port),
         {:ok, conn, ref} <- Mint.WebSocket.upgrade(websocket_scheme, conn, path, []) do
      {:ok, conn, ref}
    else
      {:error, %Mint.HTTPError{} = error} ->
        {:error, {:http_error, error.reason}}

      {:error, %Mint.TransportError{} = error} ->
        {:error, {:transport_error, error.reason}}

      {:error, conn, reason} ->
        {:error, {:websocket_upgrade_error, reason, conn}}
    end
  end

end
