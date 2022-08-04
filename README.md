# Gengo

[![GoDoc Widget](https://godoc.org/github.com/go-courier/gengo?status.svg)](https://pkg.go.dev/github.com/go-courier/gengo)
[![codecov](https://codecov.io/gh/go-courier/gengo/branch/main/graph/badge.svg)](https://codecov.io/gh/go-courier/gengo)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-courier/gengo)](https://goreportcard.com/report/github.com/go-courier/gengo)

```go
package customgen

import (
	"go/ast"
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&customGen{})
}

type customGen struct {
}

func (*customGen) Name() string {
	return "custom"
}

func (g *customGen) GenerateType(c gengo.Context, named *types.Named) error {
	if !ast.IsExported(named.Obj().Name()) {
		// skip type 
		return gengo.ErrSkip
	}

	if whenSomeThing() {
		// end generate but ignore error
		return gengo.ErrIgnore
	}
	// do generate
	return nil
}
```