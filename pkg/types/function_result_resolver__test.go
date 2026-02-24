package types

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func Test_funcResultsResolver(t *testing.T) {
	t.Run("GIVEN function with return directly", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal("(2)")),
		)
	})

	t.Run("GIVEN function with untyped nil", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal("(untyped nil)")),
		)
	})

	t.Run("GIVEN function with return multiple values directly", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal("(1, *errors.errorString)")),
		)
	})

	t.Run("GIVEN function with result assigned", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal("(2)")),
		)
	})

	t.Run("GIVEN function with return of select expr", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal(`(string | "2")`)),
		)
	})

	t.Run("GIVEN function with conditions", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal(`("a1" | "a2" | "a3", "b1" | "b2" | "b3")`)),
		)
	})

	t.Run("GIVEN function with named return", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal(`("1")`)),
		)
	})

	t.Run("GIVEN function with call", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal(`("1")`)),
		)
	})

	t.Run("GIVEN function call with inline func which return error", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal(`(*errors.errorString)`)),
		)
	})

	t.Run("GIVEN function WrapError call with first arg error", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal(`(untyped nil, *errors.errorString | untyped nil)`)),
		)
	})

	t.Run("GIVEN function with call and set to named return", func(t *testing.T) {
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

		Then(t, "get results",
			Expect(results.String(), Equal(`("1")`)),
		)
	})

	t.Run("GIVEN function with interface call", func(t *testing.T) {
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
			Then(t, "get results",
				Expect(results.String(), Equal(`(string)`)),
			)
		}

		{
			fn := p.Function("Multiple")
			results, _ := p.ResultsOf(fn)
			Then(t, "get results",
				Expect(results.String(), Equal(`(string, error)`)),
			)
		}
	})
}

func initModule(t *testing.T, pkgs map[string]string) (string, *Universe) {
	pkgPath := "x.io/test"
	pkgDir := t.TempDir()

	Must(t, func() error {
		return os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(`
module "x.io/test"

go 1.24
`), os.ModePerm)
	})

	entries := make([]string, 0, len(pkgs))

	for pkgName, pkgContent := range pkgs {
		entries = append(entries, filepath.Join(pkgPath, pkgName))

		Must(t, func() error {
			return os.MkdirAll(filepath.Join(pkgDir, pkgName), os.ModePerm)
		})

		Must(t, func() error {
			return os.WriteFile(filepath.Join(pkgDir, pkgName, pkgName+".go"), []byte(pkgContent), os.ModePerm)
		})
	}

	u := MustValue(t, func() (*Universe, error) {
		return Load(entries, WithDir(pkgDir))
	})

	t.Cleanup(func() {
		_ = os.RemoveAll(pkgDir)
	})

	return pkgPath, u
}
