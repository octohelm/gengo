module github.com/octohelm/gengo

go 1.25.3

tool github.com/octohelm/gengo/internal/cmd/fmt

// +gengo:import:group=0_controled
require github.com/octohelm/x v0.0.0-20251028032356-02d7b8d1c824

require (
	github.com/rogpeppe/go-internal v1.14.1
	golang.org/x/mod v0.29.0
	golang.org/x/text v0.31.0
	golang.org/x/tools v0.38.0
	mvdan.cc/gofumpt v0.9.2
)

require (
	github.com/go-json-experiment/json v0.0.0-20251027170946-4849db3c2f7e // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
)
