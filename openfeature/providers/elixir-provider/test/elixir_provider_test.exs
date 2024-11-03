defmodule ElixirProviderTest do
  use ExUnit.Case
  doctest ElixirProvider

  test "greets the world" do
    assert ElixirProvider.hello() == :world
  end
end
