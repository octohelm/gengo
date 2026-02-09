package format_test

import (
	"strings"
	"testing"

	"golang.org/x/mod/modfile"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/format"
)

func TestOptionsFromModFile(t *testing.T) {
	t.Run("GIVEN modfile", func(t *testing.T) {
		modContent := `
module github.com/octohelm/gengo

go 1.25.3

// +gengo:import:group=1_internal
require github.com/octohelm/x v0.0.0-20251028032356-02d7b8d1c824

// +gengo:import:group=0_controlled
require (
	x.io/a v0.1.0
	x.io/b v0.1.0
)
`

		t.Run("WHEN 解析 modfile", func(t *testing.T) {
			f := MustValue(t, func() (*modfile.File, error) {
				return modfile.Parse("", []byte(modContent), nil)
			})

			t.Run("THEN 从 modfile 获得 Options", func(t *testing.T) {
				opts := format.OptionsFromModFile(f)

				expected := format.Options{
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
				}

				Then(t, "Options应该正确",
					Expect(opts.LocalGroupPrefix, Equal(expected.LocalGroupPrefix)),
					Expect(len(opts.ImportGroups), Equal(len(expected.ImportGroups))),
				)

				t.Run("验证导入分组", func(t *testing.T) {
					group0, exists := opts.ImportGroups["0_controlled"]
					Then(t, "0_controlled分组应该存在",
						Expect(exists, Be(cmp.True())),
					)

					Then(t, "0_controlled分组前缀应该正确",
						Expect(group0.Prefixes, Equal(expected.ImportGroups["0_controlled"].Prefixes)),
					)

					group1, exists := opts.ImportGroups["1_internal"]
					Then(t, "1_internal分组应该存在",
						Expect(exists, Be(cmp.True())),
					)

					Then(t, "1_internal分组前缀应该正确",
						Expect(group1.Prefixes, Equal(expected.ImportGroups["1_internal"].Prefixes)),
					)
				})
			})
		})
	})
}

func TestSource(t *testing.T) {
	t.Run("GIVEN 源代码", func(t *testing.T) {
		source := []byte(`
package p

var ()

func f() {
	for _ = range v {
	}
}
`)

		t.Run("WHEN 格式化源代码", func(t *testing.T) {
			got, err := format.Source(source, format.Options{})

			Then(t, "格式化应该成功",
				Expect(err, Be(cmp.Nil[error]())),
				Expect(got, Be(cmp.NotZero[[]byte]())),
			)

			expected := strings.TrimSpace(`
package p

func f() {
	for range v {
	}
}
`)

			Then(t, "格式化结果应该正确",
				Expect(strings.TrimSpace(string(got)), Equal(expected)),
			)
		})
	})
}

func TestSourceWithImportsOrdered(t *testing.T) {
	t.Run("GIVEN 包含导入的源代码", func(t *testing.T) {
		source := []byte(`
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
`)

		opts := format.Options{
			LocalGroupPrefix: "x.io/a",
			ImportGroups: map[string]*format.ImportGroup{
				"b": {
					Prefixes: []string{
						"x.io/b",
					},
				},
			},
		}

		t.Run("WHEN 使用Options格式化", func(t *testing.T) {
			got, err := format.Source(source, opts)

			Then(t, "格式化应该成功",
				Expect(err, Be(cmp.Nil[error]())),
			)

			Then(t, "格式化结果应该正确",
				Expect(
					strings.TrimSpace(string(got)),
					Equal(strings.TrimSpace(`
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
`))),
			)
		})
	})

	t.Run("GIVEN 包含嵌入注释的源代码", func(t *testing.T) {
		source := []byte(`
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
`)

		t.Run("WHEN 格式化源代码", func(t *testing.T) {
			got, err := format.Source(source, format.Options{})

			Then(t, "格式化应该成功",
				Expect(err, Be(cmp.Nil[error]())),
			)

			Then(t, "格式化结果应该正确",
				Expect(
					strings.TrimSpace(string(got)),
					Equal(strings.TrimSpace(`
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
`)),
				),
			)
		})
	})
}
