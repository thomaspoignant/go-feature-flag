defmodule ElixirProvider.GoFWebSocketClient do
  use GenServer

  require Logger
  require Mint.HTTP

  alias ElixirProvider.CacheController

  @moduledoc """
  A minimal WebSocket client for listening to configuration changes from the GO Feature Flag relay proxy.
  Clears the cache on receiving change notifications.
  """

  defstruct [:conn, :websocket, :request_ref, :status, :caller, :resp_headers, :closing?]

  @type t :: %__MODULE__{
    conn: Mint.HTTP.t() | nil,
    websocket: Mint.WebSocket.t() | nil,
    request_ref: reference() | nil,
    caller: {pid(), GenServer.from()} | nil,
    status: integer() | nil,
    resp_headers: list({String.t(), String.t()}) | nil,
    closing?: boolean()
  }

  @websocket_uri "/ws/v1/flag/change"

  def connect(url) do
    with {:ok, socket} <- GenServer.start_link(__MODULE__, [], name: __MODULE__),
         {:ok, :connected} <- GenServer.call(socket, {:connect, url}) do
      {:ok, socket}
    end
  end

  def stop() do
    GenServer.stop(__MODULE__)
  end

  @impl true
  def init([]) do
    {:ok, %__MODULE__{}}
  end

  @impl true
  def handle_call({:connect, url}, from, state) do
    uri = URI.parse(url)

    http_scheme =
      case uri.scheme do
        "ws" -> :http
        "wss" -> :https
      end

    ws_scheme =
      case uri.scheme do
        "ws" -> :ws
        "wss" -> :wss
      end

    # Construct the WebSocket path
    path = uri.path <> @websocket_uri

    with {:ok, conn} <- Mint.HTTP.connect(http_scheme, uri.host, uri.port),
         {:ok, conn, ref} <- Mint.WebSocket.upgrade(ws_scheme, conn, path, []) do
      state = %{state | conn: conn, request_ref: ref, caller: from}
      {:noreply, state}
    else
      {:error, reason} ->
        {:reply, {:error, reason}, state}

      {:error, conn, reason} ->
        {:reply, {:error, reason}, put_in(state.conn, conn)}
    end
  end

  @impl GenServer
  def handle_info(message, state) do
    case Mint.WebSocket.stream(state.conn, message) do
      {:ok, conn, responses} ->
        state = put_in(state.conn, conn) |> handle_responses(responses)
        if state.closing?, do: do_close(state), else: {:noreply, state}

      {:error, conn, reason, _responses} ->
        state = put_in(state.conn, conn) |> reply({:error, reason})
        {:noreply, state}

      :unknown ->
        {:noreply, state}
    end
  end

  defp handle_responses(state, responses)

  defp handle_responses(%{request_ref: ref} = state, [{:status, ref, status} | rest]) do
    put_in(state.status, status)
    |> handle_responses(rest)
  end

  defp handle_responses(%{request_ref: ref} = state, [{:headers, ref, resp_headers} | rest]) do
    put_in(state.resp_headers, resp_headers)
    |> handle_responses(rest)
  end

  defp handle_responses(%{request_ref: ref} = state, [{:done, ref} | rest]) do
    case Mint.WebSocket.new(state.conn, ref, state.status, state.resp_headers) do
      {:ok, conn, websocket} ->
        %{state | conn: conn, websocket: websocket, status: nil, resp_headers: nil}
        |> reply({:ok, :connected})
        |> handle_responses(rest)

      {:error, conn, reason} ->
        put_in(state.conn, conn)
        |> reply({:error, reason})
    end
  end

  defp handle_responses(%{request_ref: ref, websocket: websocket} = state, [
         {:data, ref, data} | rest
       ])
       when websocket != nil do
        case Mint.WebSocket.decode(websocket, data) do
          {:ok, websocket, frames} ->
            put_in(state.websocket, websocket)
            |> handle_frames(frames)
            |> handle_responses(rest)

          {:error, websocket, reason} ->
            put_in(state.websocket, websocket)
            |> reply({:error, reason})
        end
  end

  defp handle_responses(state, [_response | rest]) do
    handle_responses(state, rest)
  end

  defp handle_responses(state, []), do: state

  def handle_frames(state, frames) do
    Enum.reduce(frames, state, fn
      {:close, _code, reason}, state ->
        Logger.debug("Closing connection: #{inspect(reason)}")
        %{state | closing?: true}

      {:text, text}, state ->

        response = Jason.decode!(text)

        case Map.get(response, "type") do
          "change" ->
            # Clear the cache when a change message is received
            CacheController.clear()
            Logger.info("Cache cleared due to configuration change notification.")

          _ -> nil
        end

        state

      frame, state ->
        Logger.debug("Unexpected frame received: #{inspect(frame)}")
        state
    end)
  end

  defp do_close(state) do
    Mint.HTTP.close(state.conn)
    Logger.info("Comfy websocket closed")
    {:stop, :normal, state}
  end

  defp reply(state, response) do
    if state.caller, do: GenServer.reply(state.caller, response)
    put_in(state.caller, nil)
  end
end
