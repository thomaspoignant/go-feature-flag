defmodule ElixirProvider.HttpClient do
  @moduledoc """
  Implements HttpClientBehaviour using Mint for HTTP requests.
  """

  # Define a struct to store HTTP connection, endpoint, and other configuration details
  defstruct [:conn, :endpoint, :headers]

  @type t :: %__MODULE__{
          conn: Mint.HTTP.t() | nil,
          endpoint: String.t(),
          headers: list()
        }

  def start_http_connection(options) do
    uri = URI.parse(options.endpoint)
    scheme = if uri.scheme == "https", do: :https, else: :http

    case Mint.HTTP.connect(scheme, uri.host, uri.port) do
      {:ok, conn} ->
        {:ok,
         %{
           conn: conn,
           endpoint: options.endpoint,
           headers: [{"content-type", "application/json"}]
         }}

      {:error, reason} ->
        {:error, reason}
    end
  end

  def post(%{conn: conn, endpoint: endpoint, headers: headers}, path, data) do
    url = URI.merge(endpoint, path) |> URI.to_string()
    body = Jason.encode!(data)

    with {:ok, conn, request_ref} <- Mint.HTTP.request(conn, "POST", url, headers, body),
         {:ok, response} <- read_response(conn, request_ref) do
      Jason.decode(response)
    else
      {:error, _conn, reason} -> {:error, reason}
      {:error, reason} -> {:error, reason}
    end
  end

  defp read_response(conn, request_ref) do
    receive do
      message ->
        case Mint.HTTP.stream(conn, message) do
          {:ok, _conn, responses} ->
            Enum.reduce_while(responses, {:ok, ""}, fn
              {:status, ^request_ref, status}, _acc ->
                if status == 200, do: {:cont, {:ok, ""}}, else: {:halt, {:error, :bad_status}}

              {:headers, ^request_ref, _headers}, acc ->
                {:cont, acc}

              {:data, ^request_ref, data}, {:ok, acc} ->
                {:cont, {:ok, acc <> data}}

              {:done, ^request_ref}, {:ok, acc} ->
                {:halt, {:ok, acc}}

              _other, acc ->
                {:cont, acc}
            end)

          :unknown ->
            {:error, :unknown_response}
        end
    after
      5_000 -> {:error, :timeout}
    end
  end
end
