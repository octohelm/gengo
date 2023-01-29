package types

import (
	"bytes"
	"go/ast"
	"go/constant"
	"go/types"
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

func (pi *pkgInfo) ResultsOf(typeFunc *types.Func) (Results, int) {
	s := typeFunc.Type().(*types.Signature)
	return pi.resolveFuncResults(s).Flatten(), s.Results().Len()
}

func (pi *pkgInfo) resolveFuncResults(s *types.Signature) (finalFuncResults Results) {
	if r, ok := pi.funcResults[s]; ok {
		return r
	}

	// registry before process to avoid stackoverflow
	pi.funcResults[s] = finalFuncResults
	defer func() {
		pi.funcResults[s] = finalFuncResults
	}()

	resultTypes := s.Results()
	n := resultTypes.Len()

	// no results
	if n == 0 {
		return nil
	}

	if node, ok := pi.signatures[s]; ok {
		switch x := node.(type) {
		case *ast.FuncDecl:
			return pi.funcResultsFrom(s, x.Type, x.Body)
		case *ast.FuncLit:
			return pi.funcResultsFrom(s, x.Type, x.Body)
		case *ast.SelectorExpr:
			if fn, ok := pi.Package.TypesInfo.Uses[x.Sel].(*types.Func); ok {
				pp := pi.u.Package(fn.Pkg().Path()).(*pkgInfo)
				switch x := pp.funcDecls[fn].(type) {
				case *ast.FuncDecl:
					return pp.funcResultsFrom(s, x.Type, x.Body)
				case *ast.FuncLit:
					return pp.funcResultsFrom(s, x.Type, x.Body)
				}
			}
		case *ast.CallExpr:
			// TODO should scan curried calls
			r := s.Results()

			finalFuncResults = make(Results, r.Len())
			for i := 0; i < r.Len(); i++ {
				finalFuncResults[i] = append(finalFuncResults[i], TypeAndValue{
					Type: r.At(i).Type(),
					At:   i,
				})
			}
			return
		}
	}

	return nil
}

func (pi *pkgInfo) funcResultsFrom(s *types.Signature, funcType *ast.FuncType, body *ast.BlockStmt) (finalFuncResults Results) {
	n := s.Results().Len()
	finalFuncResults = make(Results, n)

	if funcType == nil || body == nil {
		return nil
	}

	namedResults := make([]*ast.Ident, 0)

	for _, field := range funcType.Results.List {
		namedResults = append(namedResults, field.Names...)
	}

	variableLatestAssigns := map[*ast.Object]TypeAndValue{}

	assign := func(o *ast.Object, rhs []ast.Expr, n int, i int) {
		if len(rhs) == n {
			variableLatestAssigns[o] = TypeAndValue{Expr: rhs[i]}
		} else {
			variableLatestAssigns[o] = TypeAndValue{Expr: rhs[0], At: i}
		}
	}

	var typeAndValueOf func(at int, expr ast.Expr) TypeAndValue

	typeAndValueOf = func(at int, expr ast.Expr) (final TypeAndValue) {
		switch x := expr.(type) {
		case *ast.Ident:
			if x.Obj != nil {
				if tve, ok := variableLatestAssigns[x.Obj]; ok {
					return typeAndValueOf(tve.At, tve.Expr)
				}
			}
		case *ast.SelectorExpr:
			if x.Sel.Obj != nil {
				if tve, ok := variableLatestAssigns[x.Sel.Obj]; ok {
					return typeAndValueOf(tve.At, tve.Expr)
				}
			}
		case *ast.CallExpr:
			switch callX := pi.Package.TypesInfo.TypeOf(x.Fun).(type) {
			case *types.Signature:
				final.At = at
				final.Expr = expr

				rets := callX.Results()
				shouldDeepResolve := false

				for i := 0; i < rets.Len(); i++ {
					t := rets.At(i).Type()

					switch t.String() {
					case "error", "any", "interface{}":
						shouldDeepResolve = true
					}
				}

				if final.At < rets.Len() {
					final.Type = rets.At(final.At).Type()
				} else {
					final.Type = rets.At(0).Type()
				}

				if shouldDeepResolve {
					final.From = pi.resolveFuncResults(callX)
				}

				return final
			}
		}

		tv, _ := pi.Eval(expr)

		final.Type = tv.Type
		final.Value = tv.Value
		final.At = at
		final.Expr = expr

		return
	}

	ast.Inspect(body, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.FuncLit:
			// skip func lit
			return false
		case *ast.AssignStmt:
			// set var by code flow
			// not support side effect assign
			for i := range x.Lhs {
				switch lhs := x.Lhs[i].(type) {
				// assign to variable
				case *ast.Ident:
					if lhs.Obj != nil {
						assign(lhs.Obj, x.Rhs, len(x.Lhs), i)
					}
				// assign to field
				case *ast.SelectorExpr:
					if lhs.Sel != nil {
						assign(lhs.Sel.Obj, x.Rhs, len(x.Lhs), i)
					}
				}
			}
		case *ast.ReturnStmt:
			results := x.Results

			// fill return resolveFuncResults by named resolveFuncResults
			if x.Results == nil {
				results = make([]ast.Expr, n)

				for i := range namedResults {
					results[i] = namedResults[i]
				}
			}

			for at := 0; at < n; at++ {
				if len(results) == n {
					finalFuncResults[at] = append(finalFuncResults[at], typeAndValueOf(at, results[at]))
				} else {
					finalFuncResults[at] = append(finalFuncResults[at], typeAndValueOf(at, results[0]))
				}
			}
		}

		return true
	})

	return
}
