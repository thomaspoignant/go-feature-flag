module github.com/thomaspoignant/go-feature-flag/modules/core

go 1.25.14

require (
	github.com/diegoholiveira/jsonlogic/v3 v3.10.1
	github.com/google/go-cmp v0.7.0
	github.com/nikunjy/rules v1.5.0
	github.com/stretchr/testify v1.12.1
)

// The evaluator cache shares one evaluator per query between goroutines, which needs
// https://github.com/go-feature-flag/rules/pull/1 (compiled, concurrency-safe Process).
// TODO: point at a github.com/go-feature-flag/rules tag once that PR merges.
replace github.com/nikunjy/rules => github.com/jh125486/rules v1.5.1-0.20261116140709-a9eeb6d8d42d

require (
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20260813180055-c1d0aacb2297 // indirect
)
