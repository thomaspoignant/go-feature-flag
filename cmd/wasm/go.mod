module github.com/thomaspoignant/go-feature-flag/cmd/wasm

go 1.24.13

require (
	github.com/stretchr/testify v1.12.1
	github.com/thomaspoignant/go-feature-flag/modules/core v0.7.2
)

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
)

require (
	github.com/diegoholiveira/jsonlogic/v3 v3.10.1 // indirect
	github.com/nikunjy/rules v1.5.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
)

// TODO: remove this once https://github.com/nikunjy/rules/pull/43 merges and a new version is available
// This fix is needed to resolve semver comparison issues with prerelease versions.
// Check https://github.com/thomaspoignant/go-feature-flag/issues/4736 for more details.
replace github.com/nikunjy/rules => github.com/hairyhenderson/rules v0.0.0-20250704181428-58ee76134adc
