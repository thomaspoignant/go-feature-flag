module github.com/thomaspoignant/go-feature-flag/cmd/wasm

go 1.24.6

require github.com/thomaspoignant/go-feature-flag/modules/core v0.5.0

require (
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
)

require (
	github.com/GeorgeD19/json-logic-go v0.0.0-20220225111652-48cc2d2c387e // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/dariubs/percent v0.0.0-20190521174708-8153fcbd48ae // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/nikunjy/rules v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/stretchr/testify v1.11.1
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// TODO: remove this once https://github.com/nikunjy/rules/pull/43 merges and a new version is available
replace github.com/nikunjy/rules => github.com/hairyhenderson/rules v0.0.0-20250704181428-58ee76134adc
