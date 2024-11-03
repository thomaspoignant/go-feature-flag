defmodule ElixirProvider.MixProject do
  use Mix.Project

  def project do
    [
      app: :elixir_provider,
      version: "0.1.0",
      elixir: "~> 1.17",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
    ]
  end

  # Run "mix help compile.app" to learn about applications.
  def application do
    [
      extra_applications: [:logger]
    ]
  end

  # Run "mix help deps" to learn about dependencies.
  defp deps do
    [
      {:open_feature, git: "https://github.com/open-feature/elixir-sdk.git"},
      {:jason, "~> 1.4"},
      {:websockex, "~> 0.4.3"}
      # {:dep_from_hexpm, "~> 0.3.0"},,
      # {:dep_from_git, git: "https://github.com/elixir-lang/my_dep.git", tag: "0.1.0"}
    ]
  end
end
