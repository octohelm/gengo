package a

import (
	"context"
	"errors"
	"strings"

	"github.com/octohelm/gengo/pkg/types/testdata/a/b"
	"github.com/octohelm/gengo/pkg/types/testdata/a/x"
)

type InterfaceType interface {
	Single() string
	Multiple() (string, error)
}

func (String) Method() string {
	return strings.Join(strings.Split("1", ","), ",")
}

func FuncSingleReturn() any {
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

func FuncWillCall() (any, String) {
	return FuncSingleReturn(), String(FuncSelectExprReturn())
}

func FuncReturnWithInterfaceCall() (any, error) {
	return InterfaceType(nil).Multiple()
}

func FuncReturnWithInterfaceCallSingle() any {
	return InterfaceType(nil).Single()
}

func FuncReturnWithCallDirectly() (any, String) {
	return FuncWillCall()
}

func FuncWithNamedReturn() (a any, b String) {
	a, b = FuncWillCall()
	return
}

func FuncSingleNamedReturnByAssign() (a any, s String, err error) {
	a = "" + "1"
	s = "2"
	return a, s, errors.New("some err")
}

func FuncWithSwitch() (a any, b String) {
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

func FuncWithIf() (a any) {
	if true {
		a = "a0"
		return
	} else if true {
		a = "a1"
		return
	}
	a = strings.Join([]string{"a", "b"}, "")
	return
}

type callChain struct{}

func callChains() callChain { return callChain{} }
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

func FuncCallReturnAssign() (any, String) {
	return FuncSingleReturn(), String(FuncSelectExprReturn())
}

func FuncCallWithFuncLit() (any, String) {
	call := func() any {
		return 1
	}
	return call(), "s"
}

func FuncWithImportedCall() any {
	return b.V()
}

func FuncWithGenerics() (any, error) {
	return x.Do(context.Background(), &ListNode{})
}

type Func func() func() int

func curryCall(r bool) Func {
	if r {
		return func() func() int {
			return func() int { return 1 }
		}
	}
	return func() func() int {
		return b.V
	}
}

func FuncCurryCall() any {
	return curryCall(true)()()
}

func Tx(ctx context.Context, action func(context.Context) error) error {
	return action(ctx)
}

func FuncWithFuncArg() any {
	return Tx(context.Background(), func(ctx context.Context) error {
		if true {
			return errors.New("test 2")
		}
		return errors.New("test 1")
	})
}
