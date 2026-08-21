module github.com/octohelm/gengo

go 1.27.0

tool (
	github.com/octohelm/gengo/tool/internal/cmd/fmt
	github.com/octohelm/gengo/tool/internal/cmd/gen
	github.com/octohelm/gengo/tool/internal/cmd/skills-install
)

// +gengo:import:group=0_controlled

require (
	// +skill:testing-guideline
	github.com/octohelm/x v0.0.0-20260821032215-38f48df07f8c
	golang.org/x/mod v0.40.0
	golang.org/x/text v0.41.0
	golang.org/x/tools v0.49.0
	mvdan.cc/gofumpt v0.11.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
)
