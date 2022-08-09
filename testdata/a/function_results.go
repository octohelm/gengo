package a

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/octohelm/gengo/testdata/a/b"
)

func Example() {

}

func (String) Method() string {
	return strings.Join(strings.Split("1", ","), ",")
}

func FuncSingleReturn() any {
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

func FuncSelectExprReturn() string {
	v := struct{ s string }{}
	v.s = "2"
	return v.s
}

func FuncWillCall() (a any, s String) {
	return FuncSingleReturn(), String(FuncSelectExprReturn())
}

func FuncReturnWithCallDirectly() (a any, b String) {
	return FuncWillCall()
}

func FuncWithNamedReturn() (a any, b String) {
	a, b = FuncWillCall()
	return
}

func newErr() error {
	return errors.New("some err")
}

func FuncSingleNamedReturnByAssign() (a any, s String, err error) {
	a = "" + "1"
	s = "2"
	return a, s, newErr()
}

func FunWithSwitch() (a any, b String) {
	switch a {
	case "1":
		a = "a1"
		b = "b1"
		return
	case "2":
		a = "a2"
		b = "b2"
		return
	default:
		a = "a3"
		b = "b3"
	}
	return
}

func str(a string, b string) string {
	return a + b
}

func FuncWithIf() (a any) {
	if true {
		a = "a0"
		return
	} else if true {
		a = "a1"
		return
	} else {
		a = str("a", "b")
		return
	}
}

func callChains() callChain {
	return callChain{}
}

type callChain struct {
}

func (callChain) With() callChain {
	return callChain{}
}

func (callChain) Call() (*string, error) {
	a := ""
	return &a, nil
}

func FuncWithCallChain() (any, error) {
	var a *string

	if true {
		s, err := callChains().With().Call()
		if err != nil {
			return nil, err
		}
		a = s
	}

	return a, nil
}

func FuncCallReturnAssign() (a any, b String) {
	return FuncSingleReturn(), String(FuncSelectExprReturn())
}

func FuncCallWithFuncLit() (a any, b String) {
	call := func() any {
		return 1
	}
	return call(), "s"
}

func FuncWithImportedCall() any {
	return b.V()
}

type Func func() func() int

func curryCall(r bool) Func {
	if r {
		return func() func() int {
			return func() int {
				return 1
			}
		}
	}

	return func() func() int {
		return b.V
	}
}

func FuncCurryCall() any {
	return curryCall(true)()()
}
