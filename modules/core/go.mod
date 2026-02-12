module github.com/thomaspoignant/go-feature-flag/modules/core

go 1.24.6

require (
	github.com/diegoholiveira/jsonlogic/v3 v3.9.0
	github.com/google/go-cmp v0.7.0
	github.com/nikunjy/rules v1.5.0
	github.com/stretchr/testify v1.11.1
)

// TODO: remove this once https://github.com/nikunjy/rules/pull/43 merges and a new version is available
replace github.com/nikunjy/rules => github.com/hairyhenderson/rules v0.0.0-20250704181428-58ee76134adc

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/barkimedes/go-deepcopy v0.0.0-20220514131651-17c30cfc62df // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
