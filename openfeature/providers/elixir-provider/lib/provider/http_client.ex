defmodule ElixirProvider.HttpClient do

  @moduledoc """
  Handles HTTP requests to the GO Feature Flag API.
  """

  @type t :: Mint.HTTP.t()

  @spec start_http_connection(String.t()) :: {:ok, Mint.HTTP.t()} | {:error, any()}
  def start_http_connection(endpoint) do
    uri = URI.parse(endpoint)
    scheme = if uri.scheme == "https", do: :https, else: :http
    Mint.HTTP.connect(scheme, uri.host, uri.port)
  end

  @spec post(Mint.HTTP.t(), String.t(), map()) :: {:ok, map()} | {:error, any()}
  def post(conn, path, data) do
    headers = [{"content-type", "application/json"}]
    body = Jason.encode!(data)

    with {:ok, conn, request_ref} <- Mint.HTTP.request(conn, "POST", path, headers, body),
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
