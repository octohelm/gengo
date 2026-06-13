module github.com/octohelm/gengo

go 1.26.2

tool (
	github.com/octohelm/gengo/tool/internal/cmd/fmt
	github.com/octohelm/gengo/tool/internal/cmd/gen
	github.com/octohelm/gengo/tool/internal/cmd/skills-install
)

// +gengo:import:group=0_controlled

// +skill:testing-guideline
require github.com/octohelm/x v0.0.0-20260508104609-6b72a870e0d2

require (
	golang.org/x/mod v0.37.0
	golang.org/x/text v0.38.0
	golang.org/x/tools v0.46.0
	mvdan.cc/gofumpt v0.10.0
)

require (
	github.com/go-json-experiment/json v0.0.0-20260214004413-d219187c3433 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
)
