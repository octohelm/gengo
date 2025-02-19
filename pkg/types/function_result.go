package types

import (
	"bytes"
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
)

type Results []TypeAndValues

func (results Results) Flatten() Results {
	finalResults := make(Results, len(results))

	for at := range finalResults {
		typeAndValues := results[at]

		for i := range typeAndValues {
			e := typeAndValues[i]

			if e.From != nil {
				fromRets := e.From.Flatten()

				// _, _, err = x, y, fromRets[0]
				ret := fromRets[0]

				if e.At < len(fromRets) {
					ret = fromRets[e.At]
				}

				finalResults[at] = append(finalResults[at], ret...)
			} else {
				finalResults[at] = append(finalResults[at], e)
			}
		}
	}

	return finalResults
}

func (results Results) String() string {
	buf := bytes.NewBuffer(nil)

	buf.WriteString("(")

	for i := range results {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(results[i].String())
	}

	buf.WriteString(")")

	return buf.String()
}

func (results Results) Concat(result2 Results) (finalResult Results) {
	if len(results) == len(result2) {
		for i, values := range result2 {
			results[i] = slices.Concat(results[i], values)
		}
	}
	return results
}

type TypeAndValues []TypeAndValue

func (typeOrValues TypeAndValues) String() string {
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

type TypeAndValue struct {
	// Type
	Type types.Type
	// Value not nil if can be static value
	Value constant.Value
	// Expr TypeOrValue assigned by Expr
	Expr ast.Expr
	// if Expr == *ast.CallExpr, will use this to pick value
	At int
	// if Expr == *ast.CallExpr, record all available results
	From Results
}

func (r TypeAndValue) String() string {
	if r.Value != nil {
		return r.Value.String()
	} else if r.Type != nil {
		return r.Type.String()
	} else if len(r.From) != 0 {
		return r.From[r.At].String()
	}
	return "invalid"
}
