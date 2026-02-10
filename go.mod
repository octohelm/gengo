module github.com/octohelm/gengo

go 1.25.6

tool github.com/octohelm/gengo/internal/cmd/fmt

// +gengo:import:group=0_controlled
require github.com/octohelm/x v0.0.0-20260209062350-0af9cf4d4286

require (
	golang.org/x/mod v0.33.0
	golang.org/x/text v0.33.0
	golang.org/x/tools v0.41.0
	mvdan.cc/gofumpt v0.9.2
)

require (
	github.com/go-json-experiment/json v0.0.0-20251027170946-4849db3c2f7e // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
)
