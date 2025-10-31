package format_test

import (
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/modfile"

	testingx "github.com/octohelm/x/testing"
	"github.com/octohelm/x/testing/bdd"

	"github.com/octohelm/gengo/pkg/format"
)

func TestOptionsFromModFile(t *testing.T) {
	f, _ := modfile.Parse("", []byte(`
module github.com/octohelm/gengo

go 1.25.3

// +gengo:import:group=1_internal
require github.com/octohelm/x v0.0.0-20251028032356-02d7b8d1c824

// +gengo:import:group=0_controlled
require (
	x.io/a v0.1.0
	x.io/b v0.1.0
)
`), nil)

	testingx.Expect(t, format.OptionsFromModFile(f), testingx.Equal(format.Options{
		LocalGroupPrefix: "github.com/octohelm/gengo",
		ImportGroups: map[string]*format.ImportGroup{
			"0_controlled": {
				Prefixes: []string{
					"x.io/a",
					"x.io/b",
				},
			},
			"1_internal": {
				Prefixes: []string{
					"github.com/octohelm/x",
				},
			},
		},
	}))
}

func TestSource(t *testing.T) {
	bdd.FromT(t).When("do format", func(b bdd.T) {
		got, err := format.Source([]byte(`
package p

var ()

func f() {
	for _ = range v {
	}
}
`), format.Options{})

		b.Then("success",
			bdd.NoError(err),
			bdd.Equal(strings.TrimSpace(`
package p

func f() {
	for range v {
	}
}
`), strings.TrimSpace(string(got))),
		)
	})
}

func TestSourceWithImportsOrdered(t *testing.T) {
	b := bdd.FromT(t)

	b.When("do format", func(b bdd.T) {
		got, err := format.Source([]byte(`
// pkg comment
package p

import (
	_ "embed"
	_ "x.io/a/pkg/side"
	_ "x.io/b/pkg/side"
	_ "x.io/c/pkg/side"
	z "x.io/c/pkg/z"
	"strings"
	"x.io/b/pkg/y"
	"x.io/a/pkg/x"
)

import "C"

// other comment
var (
	X = ""
)

func f() {
	_ = strings.TrimSpace("")
	_ = x.X()
	_ = y.Y()
	_ = z.Z()
}
`), format.Options{
			LocalGroupPrefix: "x.io/a",
			ImportGroups: map[string]*format.ImportGroup{
				"b": {
					Prefixes: []string{
						"x.io/b",
					},
				},
			},
		})

		b.Then("success",
			bdd.NoError(err),
			bdd.Equal(strings.TrimSpace(`
// pkg comment
package p

import (
	"strings"

	z "x.io/c/pkg/z"

	"x.io/b/pkg/y"

	"x.io/a/pkg/x"
)

import (
	_ "embed"

	_ "x.io/c/pkg/side"

	_ "x.io/b/pkg/side"

	_ "x.io/a/pkg/side"
)

import "C"

// other comment
var (
	X = ""
)

func f() {
	_ = strings.TrimSpace("")
	_ = x.X()
	_ = y.Y()
	_ = z.Z()
}
`), strings.TrimSpace(string(got))),
		)
	})

	b.Given("side embed", func(b bdd.T) {
		b.When("do format", func(b bdd.T) {
			got, err := format.Source([]byte(`
package p

import (
	_ "embed"
	"strings"
)

//go:embed x.json
var X []byte

func f() {
	_ = strings.TrimSpace("")
}
`), format.Options{})

			b.Then("success",
				bdd.NoError(err),
				bdd.Equal(strings.TrimSpace(`
package p

import (
	"strings"
)

import (
	_ "embed"
)

//go:embed x.json
var X []byte

func f() {
	_ = strings.TrimSpace("")
}
`), strings.TrimSpace(string(got))),
			)
		})
	})
}
