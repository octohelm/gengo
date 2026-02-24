module github.com/octohelm/gengo

go 1.26.0

tool github.com/octohelm/gengo/internal/cmd/fmt

// +gengo:import:group=0_controlled
require github.com/octohelm/x v0.0.0-20260224021741-724787200747

require (
	golang.org/x/mod v0.33.0
	golang.org/x/text v0.34.0
	golang.org/x/tools v0.42.0
	mvdan.cc/gofumpt v0.9.2
)

require (
	github.com/go-json-experiment/json v0.0.0-20260214004413-d219187c3433 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
)
