package types

import (
	"go/ast"
	"go/token"
	"go/types"
	"iter"
)

func (p *pkgInfo) ResultsOf(typeFunc *types.Func) (results FuncResults, n int) {
	s := typeFunc.Type().(*types.Signature)
	r := p.funcResultsResolverFor(s)

	return r.Results(visits{}), r.Len()
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
type visits map[*ast.FuncType][]bool

func (v visits) visited(t *ast.FuncType, at int) bool {
	if n, ok := v[t]; ok {
		return n[at]
	}

	n := 0
	for _, f := range t.Results.List {
		if len(f.Names) > 0 {
			n += len(f.Names)
		} else {
			n += 1
		}
	}

	v[t] = make([]bool, n)
	v[t][at] = true
	return false
}

func (r *funcResultsResolver) resolverFor(p1 *pkgInfo, sig *types.Signature) *funcResultsResolver {
	return &funcResultsResolver{
		pkgInfo: p1,
		sig:     sig,
	}
}

func (r *funcResultsResolver) Len() int {
	return r.sig.Results().Len()
}

func (r *funcResultsResolver) Results(vs visits) (finalFuncResults FuncResults) {
	retN := r.sig.Results().Len()

	// no results
	if retN == 0 {
		return finalFuncResults
	}

	if node, ok := r.signatures[r.sig]; ok {
		switch x := node.(type) {
		case *ast.FuncDecl:
			return r.resultsFromAst(vs, x.Type, x.Body)
		case *ast.FuncLit:
			return r.resultsFromAst(vs, x.Type, x.Body)
		case *ast.SelectorExpr:
			if fn, ok := r.Package.TypesInfo.Uses[x.Sel].(*types.Func); ok {
				results := funcResultsFromSignature(r.sig)

				resolveInFunc := func() FuncResults {
					p1 := r.u.Package(fn.Pkg().Path()).(*pkgInfo)

					switch x := p1.funcDecls[fn].(type) {
					case *ast.FuncDecl:
						return r.resolverFor(p1, fn.Signature()).resultsFromAst(vs, x.Type, x.Body)
					case *ast.FuncLit:
						return r.resolverFor(p1, fn.Signature()).resultsFromAst(vs, x.Type, x.Body)
					}

					return FuncResults{}
				}

				return results.Concat(resolveInFunc())
			}
		case *ast.CallExpr:
			// TODO should scan curried
			return funcResultsFromSignature(r.sig)
		}
	} else {
		// interface without ast found
		return funcResultsFromSignature(r.sig)
	}

	return finalFuncResults
}

func funcResultsFromSignature(sig *types.Signature) FuncResults {
	rets := sig.Results()
	finalFuncResults := make(FuncResults, rets.Len())
	for i := 0; i < rets.Len(); i++ {
		finalFuncResults[i] = append(finalFuncResults[i], Result{
			Type: rets.At(i).Type(),
		})
	}
	return finalFuncResults
}

func (r *funcResultsResolver) resultsAt(vs visits, at int) iter.Seq[Result] {
	retN := r.sig.Results().Len()

	// no results or out of range
	if retN == 0 && at >= retN {
		return func(yield func(Result) bool) {
		}
	}

	if node, ok := r.signatures[r.sig]; ok {
		switch x := node.(type) {
		case *ast.FuncDecl:
			return r.resultsFromAstAt(vs, at, x.Type, x.Body)
		case *ast.FuncLit:
			return r.resultsFromAstAt(vs, at, x.Type, x.Body)
		case *ast.SelectorExpr:
			if fn, ok := r.Package.TypesInfo.Uses[x.Sel].(*types.Func); ok {
				p1 := r.u.Package(fn.Pkg().Path()).(*pkgInfo)

				switch x := p1.funcDecls[fn].(type) {
				case *ast.FuncDecl:
					return r.resolverFor(p1, fn.Signature()).resultsFromAstAt(vs, at, x.Type, x.Body)
				case *ast.FuncLit:
					return r.resolverFor(p1, fn.Signature()).resultsFromAstAt(vs, at, x.Type, x.Body)
				}
			}
		}
	}

	return func(yield func(Result) bool) {
	}
}

func (r *funcResultsResolver) resultsFromAst(vs visits, funcType *ast.FuncType, body *ast.BlockStmt) FuncResults {
	if funcType == nil {
		return nil
	}

	finalResults := make(FuncResults, r.Len())

	for at := 0; at < r.Len(); at++ {
		for tv := range r.resultsFromAstAt(vs, at, funcType, body) {
			finalResults[at] = append(finalResults[at], tv)
		}

		if len(finalResults[at]) == 0 {
			finalResults[at] = append(finalResults[at], Result{
				Type: r.sig.Results().At(at).Type(),
			})
		}
	}

	return finalResults
}

func (r *funcResultsResolver) resultsFromAstAt(vs visits, at int, funcType *ast.FuncType, body *ast.BlockStmt) iter.Seq[Result] {
	if funcType == nil || body == nil {
		return func(yield func(Result) bool) {
		}
	}

	// avoid loop
	if ok := vs.visited(funcType, at); ok {
		return func(yield func(Result) bool) {
		}
	}

	returnStmts := func(body ast.Node) iter.Seq[*ast.ReturnStmt] {
		return func(yield func(*ast.ReturnStmt) bool) {
			if body == nil {
				return
			}

			emit := func(stmt *ast.ReturnStmt) bool {
				return yield(stmt)
			}

			ast.Inspect(body, func(node ast.Node) bool {
				if node == nil {
					return false
				}
				switch x := node.(type) {
				case *ast.FuncLit:
					// skip func lit
					return false
				case *ast.ReturnStmt:
					if !emit(x) {
						return false
					}
				}
				return true
			})
		}
	}

	return func(yield func(Result) bool) {
		for returnStmt := range returnStmts(body) {
			// named result
			if returnStmt.Results == nil {
				target := r.namedResultObjectAt(funcType, at)
				if target != nil {
					for ret := range r.assignedResultsUntil(vs, target, body, returnStmt.End()) {
						if !yield(Result{
							Type:  ret.Type,
							Value: ret.Value,
							Expr:  ret.Expr,
						}) {
							return
						}
					}
				}
				continue
			}

			for ret := range r.resultsAtReturnOrAssignment(vs, returnStmt.Results, r.Len(), at) {
				switch x := ret.Expr.(type) {
				case *ast.SelectorExpr:
					if x.Sel.Obj == nil {
						if !yield(ret) {
							return
						}
					}

					target := r.Package.TypesInfo.ObjectOf(x.Sel)

					for ret := range r.assignedResultsUntil(vs, target, body, x.Pos()) {
						if !yield(ret) {
							return
						}
					}
				case *ast.Ident:
					if x.Obj == nil {
						if !yield(ret) {
							return
						}
					}

					target := r.Package.TypesInfo.ObjectOf(x)
					for assigned := range r.assignedResultsUntil(vs, target, body, x.Pos()) {
						if !yield(assigned) {
							return
						}
					}
				default:
					if !yield(ret) {
						return
					}
				}
			}
		}
	}
}

func (r *funcResultsResolver) resultsAtReturnOrAssignment(vs visits, rhs []ast.Expr, retN int, at int) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		if n := len(rhs); n < retN && n > 0 {
			switch x := rhs[0].(type) {
			case *ast.CallExpr:
				for ret := range r.callExprResultAt(vs, at, x) {
					if !yield(ret) {
						return
					}
				}
			}
			return
		}

		for retAt, expr := range rhs {
			if retAt != at {
				continue
			}
			switch x := expr.(type) {
			case *ast.CallExpr:
				for ret := range r.callExprResultAt(vs, 0, x) {
					if !yield(ret) {
						return
					}
				}
			default:
				// value direct
				ret, _ := r.Eval(expr)

				if !yield(Result{
					Type:  ret.Type,
					Value: ret.Value,
					Expr:  expr,
				}) {
					return
				}
			}
		}
	}
}

func (r *funcResultsResolver) callExprResultAt(vs visits, at int, callExpr *ast.CallExpr) iter.Seq[Result] {
	return func(yield func(Result) bool) {
		switch fn := r.Package.TypesInfo.TypeOf(callExpr.Fun).(type) {
		case *types.Signature:
			rets := fn.Results()

			for retAt := 0; retAt < rets.Len(); retAt++ {
				if retAt == at {
					retType := rets.At(retAt).Type()
					retTypeStr := retType.String()

					switch retTypeStr {
					case "error", "any", "interface{}":
						if retTypeStr == "error" {
							for i, arg := range callExpr.Args {
								if i < fn.Params().Len() {
									argType := fn.Params().At(i)

									if argType.Type().String() == "error" {
										for ret := range r.resultsAtReturnOrAssignment(vs, []ast.Expr{arg}, 1, 0) {
											if !yield(ret) {
												return
											}
										}
									}
								}

								switch x := arg.(type) {
								case *ast.FuncLit:
									switch inlineFn := r.Package.TypesInfo.TypeOf(x).(type) {
									case *types.Signature:
										inlineFnRets := inlineFn.Results()

										for inlineRetAt := 0; inlineRetAt < inlineFnRets.Len(); inlineRetAt++ {
											if inlineRetType := rets.At(inlineRetAt).Type(); inlineRetType.String() == "error" {
												for ret := range r.resultsFromAstAt(vs, inlineRetAt, x.Type, x.Body) {
													if !yield(ret) {
														return
													}
												}
											}
										}
									}
								}
							}
						}

						resolver := r.pkgInfo.funcResultsResolverFor(fn)

						for ret := range resolver.resultsAt(vs, at) {
							if !yield(ret) {
								return
							}
						}
					default:
						if !yield(Result{Type: retType, Expr: callExpr}) {
							return
						}
					}
				}
			}
		}
	}
}

func (r *funcResultsResolver) assignedResultsUntil(vs visits, target types.Object, body ast.Node, until token.Pos) iter.Seq[Result] {
	sourceResults := iter.Seq[Result](func(yield func(Result) bool) {
	})

	if body == nil {
		return sourceResults
	}

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		if node.Pos() >= until {
			return false
		}

		switch x := node.(type) {
		case *ast.FuncLit:
			// skip func lit
			return false
		case *ast.AssignStmt:
			for i := range x.Lhs {
				switch lhs := x.Lhs[i].(type) {
				// assign to variable
				case *ast.Ident:
					if r.Package.TypesInfo.ObjectOf(lhs) == target {
						sourceResults = r.resultsAtReturnOrAssignment(vs, x.Rhs, len(x.Lhs), i)
					}
				// assign to field
				case *ast.SelectorExpr:
					if r.Package.TypesInfo.ObjectOf(lhs.Sel) == target {
						sourceResults = r.resultsAtReturnOrAssignment(vs, x.Rhs, len(x.Lhs), i)
					}
				}
			}
			return false
		}

		return true
	})

	return sourceResults
}

func (r *funcResultsResolver) namedResultObjectAt(funcType *ast.FuncType, at int) types.Object {
	retAt := 0
	for _, field := range funcType.Results.List {
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				if retAt == at {
					return r.Package.TypesInfo.ObjectOf(name)
				}
				retAt++
			}
		} else {
			retAt++
		}
	}
	return nil
}
