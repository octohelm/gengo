package types

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func Test_funcResultsResolver(t *testing.T) {
	t.Run("GIVEN function with return directly", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

func Fn() any {
	return 2
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))
		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal("(2)", results.String()),
		)
	}))

	t.Run("GIVEN function with untyped nil", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

func Fn() any {
	return nil
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))
		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal("(untyped nil)", results.String()),
		)
	}))

	t.Run("GIVEN function with return multiple values directly", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

import "errors"

func Fn() (any, error) {
	return 1, errors.New("test")
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))
		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal("(1, *errors.errorString)", results.String()),
		)
	}))

	t.Run("GIVEN function with result assigned", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

func Fn() any {
	// should skip
	_ = func() bool {
		a := true
		return !a
	}()

	var a any
	a = "" + "1"
	a = 2

	return a
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))
		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal("(2)", results.String()),
		)
	}))

	t.Run("GIVEN function with return of select expr", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

func Fn() any {
	v := struct{ s string }{}
	v.s = "2"
	return v.s
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))
		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal(`(string | "2")`, results.String()),
		)
	}))

	t.Run("GIVEN function with conditions", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

type String string

func Fn() (a any, b String) {
	switch a {
	case "1":
		a = "a1"
		b = "b1"
		return
	case "2":
		a = "a2"
		b = "b2"
		return
	}
	if true {
		a = "a3"
		b = "b3"
	}
	return
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))
		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal(`("a1" | "a2" | "a3", "b1" | "b2" | "b3")`, results.String()),
		)
	}))

	t.Run("GIVEN function with named return", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

func Fn() (ret string) {
	ret = "1"
	return
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))
		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal(`("1")`, results.String()),
		)
	}))

	t.Run("GIVEN function with call", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

func fn() any {
	return "1"
}

func Fn() (any) {
	return fn()
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))

		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal(`("1")`, results.String()),
		)
	}))

	t.Run("GIVEN function call with inline func which return error", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

import "errors"

func call(action func() error) error {
	return action()
}

func Fn() error {
	return call(func() error {
		return errors.New("error")
	})
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))

		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal(`(*errors.errorString)`, results.String()),
		)
	}))

	t.Run("GIVEN function WrapError call with first arg error", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

import "errors"

func Wrap[T any](x T) T {
	return x
}

func WrapError(err error) error {
	return err
}

func call(action func() error) error {
	return action()
}

func action() error {
	return call(func() error {
		return errors.New("error")
	})
}

func Fn() (any, error) {
	if e := action(); e != nil {
		return nil, WrapError(e)
	}
	return Wrap[any](nil), nil
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))

		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal(`(untyped nil, *errors.errorString | untyped nil)`, results.String()),
		)
	}))

	t.Run("GIVEN function with call and set to named return", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

func fn() any {
	return "1"
}

func Fn() (a any) {
	a = fn()
	return 
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))

		fn := p.Function("Fn")

		results, _ := p.ResultsOf(fn)

		b.Then("get results",
			bdd.Equal(`("1")`, results.String()),
		)
	}))

	t.Run("GIVEN function with interface call", bdd.GivenT(func(b bdd.T) {
		root, u := initModule(t, map[string]string{
			"a": `
package a

type InterfaceType interface {
	Single() string
	Multiple() (string, error)
}

func Single() (any) {
	return InterfaceType(nil).Single()
}

func Multiple() (any, error) {
	return InterfaceType(nil).Multiple()
}
`,
		})

		p := u.Package(filepath.Join(root, "a"))

		{
			fn := p.Function("Single")
			results, _ := p.ResultsOf(fn)
			b.Then("get results",
				bdd.Equal(`(string)`, results.String()),
			)
		}

		{
			fn := p.Function("Multiple")
			results, _ := p.ResultsOf(fn)
			b.Then("get results",
				bdd.Equal(`(string, error)`, results.String()),
			)
		}
	}))
}

func initModule(t *testing.T, pkgs map[string]string) (string, *Universe) {
	pkgPath := "x.io/test"
	pkgDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(`
module "x.io/test"

go 1.24
`), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	entries := make([]string, 0, len(pkgs))

	for pkgName, pkgContent := range pkgs {
		entries = append(entries, filepath.Join(pkgPath, pkgName))

		if err := os.MkdirAll(filepath.Join(pkgDir, pkgName), os.ModePerm); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(pkgDir, pkgName, pkgName+".go"), []byte(pkgContent), os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}

	u := bdd.Must(Load(entries, WithDir(pkgDir)))

	t.Cleanup(func() {
		_ = os.RemoveAll(pkgDir)
	})

	return pkgPath, u
}
