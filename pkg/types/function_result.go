package types

import (
	"bytes"
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
)

type FuncResults []Results

func (funcResults FuncResults) String() string {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("(")

	for i := range funcResults {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(funcResults[i].String())
	}

	buf.WriteString(")")

	return buf.String()
}

func (funcResults FuncResults) Concat(funcResults2 FuncResults) (finalFuncResults FuncResults) {
	if len(finalFuncResults) == len(funcResults2) {
		for i, results := range funcResults2 {
			funcResults[i] = slices.Concat(funcResults[i], results)
		}
	}
	return finalFuncResults
}

// TypeAndValues
// Deprecated use Results directly
type TypeAndValues = Results

type Results []Result

func (typeOrValues Results) String() string {
	buf := bytes.NewBuffer(nil)

	for i := range typeOrValues {
		r := typeOrValues[i]
		if i != 0 {
			buf.WriteString(" | ")
		}
		buf.WriteString(r.String())
	}

	return buf.String()
}

type Result struct {
	Value constant.Value
	Type  types.Type
	Expr  ast.Expr
}

func (r Result) String() string {
	if r.Value != nil {
		return r.Value.String()
	}
	if r.Type != nil {
		return r.Type.String()
	}
	return "invalid"
}
