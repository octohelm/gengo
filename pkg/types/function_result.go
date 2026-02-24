package types

import (
	"bytes"
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
)

// FuncResults 保存函数每个返回位可能出现的结果集合。
type FuncResults []Results

// String 以便于调试和测试的形式格式化推导结果。
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

// Concat 按返回位合并另一个 FuncResults。
func (funcResults FuncResults) Concat(funcResults2 FuncResults) (finalFuncResults FuncResults) {
	if len(funcResults) != len(funcResults2) {
		return funcResults
	}

	finalFuncResults = make(FuncResults, len(funcResults))
	for i, results := range funcResults {
		finalFuncResults[i] = slices.Concat(results, funcResults2[i])
	}

	return finalFuncResults
}

// TypeAndValues 保留为 Results 的废弃别名。
//
// Deprecated: 直接使用 Results。
type TypeAndValues = Results

// Results 保存单个返回位可能出现的值。
type Results []Result

// String 格式化单个返回位的候选值。
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

// Result 描述某个返回位推导出的一个可能值。
type Result struct {
	Value constant.Value
	Type  types.Type
	Expr  ast.Expr
}

// String 输出当前结果可用的最具体表示形式。
func (r Result) String() string {
	if r.Value != nil {
		return r.Value.String()
	}
	if r.Type != nil {
		return r.Type.String()
	}
	return "invalid"
}
