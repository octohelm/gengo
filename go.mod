module github.com/octohelm/gengo

go 1.26.2

tool github.com/octohelm/gengo/tool/internal/cmd/fmt

// +gengo:import:group=0_controlled
require github.com/octohelm/x v0.0.0-20260421082716-a77c6918d9d0

require (
	golang.org/x/mod v0.35.0
	golang.org/x/text v0.36.0
	golang.org/x/tools v0.44.0
	mvdan.cc/gofumpt v0.9.2
)

require (
	github.com/go-json-experiment/json v0.0.0-20260214004413-d219187c3433 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
)
