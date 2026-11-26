module github.com/thomaspoignant/go-feature-flag/modules/core

go 1.25.14

require (
	github.com/diegoholiveira/jsonlogic/v3 v3.10.3
	github.com/google/go-cmp v0.7.0
	github.com/nikunjy/rules v1.5.0
	github.com/stretchr/testify v1.12.1
)

// github.com/go-feature-flag/rules is our maintained fork of github.com/nikunjy/rules (slow to merge changes).
// It adds proper semver support (prerelease comparison, see https://github.com/thomaspoignant/go-feature-flag/issues/4736)
// and rejects invalid rule syntax at parse time.
replace github.com/nikunjy/rules => github.com/go-feature-flag/rules v0.0.0-20261118103042-e9704fa7a06c

require (
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20260813180055-c1d0aacb2297 // indirect
)
