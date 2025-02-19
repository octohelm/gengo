package types

import (
	"go/ast"
	"go/types"
	"sync"
)

func (p *pkgInfo) ResultsOf(typeFunc *types.Func) (results Results, n int) {
	s := typeFunc.Type().(*types.Signature)
	r := p.funcResultsResolverFor(s)
	
	return r.Results(visits{}).Flatten(), r.Len()
}

func (p *pkgInfo) funcResultsResolverFor(sig *types.Signature) *funcResultsResolver {
	return &funcResultsResolver{
		pkgInfo: p,
		sig:     sig,
	}
}

type funcResultsResolver struct {
	*pkgInfo
	sig *types.Signature
}
type visits map[*ast.FuncType]bool

func (v visits) visited(t *ast.FuncType) bool {
	if _, ok := v[t]; ok {
		return true
	}
	v[t] = true
	return false
}

func (r *funcResultsResolver) resolverFor(p1 *pkgInfo, sig *types.Signature) *funcResultsResolver {
	r1 := &funcResultsResolver{
		pkgInfo: p1,
		sig:     sig,
	}
	return r1
}

func (r *funcResultsResolver) Len() int {
	return r.sig.Results().Len()
}

func (r *funcResultsResolver) Results(vs visits) (finalFuncResults Results) {
	retN := r.sig.Results().Len()

	// no results
	if retN == 0 {
		return nil
	}

	if node, ok := r.signatures[r.sig]; ok {
		switch x := node.(type) {
		case *ast.FuncDecl:
			return r.fromAstOnce(vs, x.Type, x.Body)
		case *ast.FuncLit:
			return r.fromAstOnce(vs, x.Type, x.Body)
		case *ast.SelectorExpr:
			if fn, ok := r.Package.TypesInfo.Uses[x.Sel].(*types.Func); ok {
				results := funcResultsFromSignature(r.sig)

				resolveInFunc := func() Results {
					p1 := r.u.Package(fn.Pkg().Path()).(*pkgInfo)

					switch x := p1.funcDecls[fn].(type) {
					case *ast.FuncDecl:
						return r.resolverFor(p1, fn.Signature()).fromAstOnce(vs, x.Type, x.Body)
					case *ast.FuncLit:
						return r.resolverFor(p1, fn.Signature()).fromAstOnce(vs, x.Type, x.Body)
					}
					return Results{}
				}

				return results.Concat(resolveInFunc())
			}
		case *ast.CallExpr:
			// TODO should scan curried calls
			finalFuncResults = funcResultsFromSignature(r.sig)
			return
		}
	} else {
		// interface without ast found
		rets := r.sig.Results()

		finalFuncResults = make(Results, rets.Len())
		for i := 0; i < rets.Len(); i++ {
			finalFuncResults[i] = append(finalFuncResults[i], TypeAndValue{
				Type: rets.At(i).Type(),
				At:   i,
			})
		}
	}

	return nil
}

func (r *funcResultsResolver) fromAstOnce(vs visits, funcType *ast.FuncType, body *ast.BlockStmt) Results {
	if funcType == nil {
		return nil
	}

	// avoid loop
	if ok := vs.visited(funcType); ok {
		return nil
	}

	get, _ := r.funcResults.LoadOrStore(funcType, sync.OnceValue(func() Results {
		return r.fromAst(vs, funcType, body)
	}))

	return get.(func() Results)()
}

func (r *funcResultsResolver) fromAst(vs visits, funcType *ast.FuncType, body *ast.BlockStmt) (finalFuncResults Results) {
	if funcType == nil || body == nil {
		return nil
	}

	retN := r.Len()
	finalFuncResults = make(Results, retN)
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
			switch callX := r.Package.TypesInfo.TypeOf(x.Fun).(type) {
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
					final.From = r.pkgInfo.funcResultsResolverFor(callX).Results(vs)
				}

				return final
			}
		}

		tv, _ := r.Eval(expr)

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
				results = make([]ast.Expr, retN)

				for i := range namedResults {
					results[i] = namedResults[i]
				}
			}

			for at := 0; at < retN; at++ {
				if len(results) == retN {
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

func funcResultsFromSignature(sig *types.Signature) Results {
	rets := sig.Results()
	finalFuncResults := make(Results, rets.Len())
	for i := 0; i < rets.Len(); i++ {
		finalFuncResults[i] = append(finalFuncResults[i], TypeAndValue{
			Type: rets.At(i).Type(),
			At:   i,
		})
	}
	return finalFuncResults
}
