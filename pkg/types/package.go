package types

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"sync"

	"github.com/octohelm/x/ptr"
	"golang.org/x/tools/go/packages"
)

type Module struct {
	packages.Module
}

type Package interface {
	// Pkg of go package
	Pkg() *types.Package
	// Imports of go package
	Imports() map[string]Package
	// Module of go package
	Module() *packages.Module
	// SourceDir code source absolute dir
	SourceDir() string
	// FileSet return file set when load
	FileSet() *token.FileSet
	// Files ast files of package
	Files() []*ast.File
	// Decl ast Decl for pos
	Decl(pos token.Pos) ast.Decl
	// Doc comment tags and leading comments for pos
	Doc(pos token.Pos) (map[string][]string, []string)
	// Comment trailing comments for pos
	Comment(pos token.Pos) []string
	// Eval eval expr in package
	Eval(expr ast.Expr) (types.TypeAndValue, error)
	// Constant get constant by name
	Constant(name string) *types.Const
	// Constants get all constants of package
	Constants() map[string]*types.Const
	// Type get type by name
	Type(name string) *types.TypeName
	// Types get all types of package
	Types() map[string]*types.TypeName
	// Function get function by name
	Function(name string) *types.Func
	// Functions get all signatures of package
	Functions() map[string]*types.Func
	// MethodsOf get methods of types.TypeName
	MethodsOf(n *types.Named, canPtr bool) []*types.Func
	// ResultsOf get possible TypeAndValue of function
	ResultsOf(tpe *types.Func) (results Results, resultN int)
	// Position get position of pos
	Position(pos token.Pos) token.Position
	// ObjectOf get object of ident
	ObjectOf(id *ast.Ident) types.Object
}

type ModInfo struct {
	Module  string
	Require map[string]string
}

func newPkg(pkg *packages.Package, u *Universe) Package {
	p := &pkgInfo{
		u:       u,
		Package: pkg,

		imports: make(map[string]Package),

		constants: make(map[string]*types.Const),
		types:     make(map[string]*types.TypeName),
		funcs:     make(map[string]*types.Func),
		methods:   make(map[*types.Named][]*types.Func),

		endLineToCommentGroup:         make(map[fileLine]*ast.CommentGroup),
		endLineToTrailingCommentGroup: make(map[fileLine]*ast.CommentGroup),

		funcDecls:  make(map[*types.Func]ast.Node),
		signatures: make(map[*types.Signature]ast.Node),
	}

	for pkgPath := range pkg.Imports {
		p.imports[pkgPath] = u.Package(pkgPath)
	}

	fileLineFor := func(pos token.Pos, deltaLine int) fileLine {
		position := p.Package.Fset.Position(pos)
		return fileLine{position.Filename, position.Line + deltaLine}
	}

	collectCommentGroup := func(c *ast.CommentGroup, isTrailing bool, stmtPos token.Pos) {
		fl := fileLineFor(stmtPos, 0)

		if c != nil && c.Pos() == stmtPos {
			// stmt is CommentGroup
			fl = fileLineFor(c.End(), 0)
		} else if !isTrailing {
			fl = fileLineFor(stmtPos, -1)
		}

		if isTrailing {
			if cc := p.endLineToTrailingCommentGroup[fl]; cc == nil {
				p.endLineToTrailingCommentGroup[fl] = c
			}
		} else {
			if cc := p.endLineToCommentGroup[fl]; cc == nil {
				p.endLineToCommentGroup[fl] = c
			}
		}
	}

	for ident := range p.Package.TypesInfo.Defs {
		switch x := p.Package.TypesInfo.Defs[ident].(type) {
		case *types.Func:
			s := x.Type().(*types.Signature)

			if r := s.Recv(); r != nil {
				var named *types.Named

				switch t := r.Type().(type) {
				case *types.Pointer:
					if n, ok := t.Elem().(*types.Named); ok {
						named = n
					}
				case *types.Named:
					named = t
				}

				if named != nil {
					p.methods[named] = append(p.methods[named], x)
				}
			} else {
				p.funcs[x.Name()] = x
			}
		case *types.TypeName:
			p.types[x.Name()] = x
		case *types.Const:
			p.constants[x.Name()] = x
		}
	}

	for i := range p.Package.Syntax {
		f := p.Package.Syntax[i]

		ast.Inspect(f, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				fn := p.Package.TypesInfo.TypeOf(x.Name)
				if fn != nil {
					if s, ok := fn.(*types.Signature); ok {
						if _, ok := p.signatures[s]; !ok {
							p.signatures[s] = x
						}
					}
					if f, ok := p.Package.TypesInfo.ObjectOf(x.Name).(*types.Func); ok {
						p.funcDecls[f] = x
					}
				}
			case *ast.FuncLit:
				fn := p.Package.TypesInfo.TypeOf(x.Type)
				if fn != nil {
					if s, ok := fn.(*types.Signature); ok {
						if _, ok := p.signatures[s]; !ok {
							p.signatures[s] = x
						}
					}
					for _, o := range p.Package.TypesInfo.Defs {
						if f, ok := o.(*types.Func); ok {
							if f.Pos() == x.Pos() {
								p.funcDecls[f] = x
							}
						}
					}
				}
			case *ast.CommentGroup:
				collectCommentGroup(x, false, x.Pos())
			case *ast.ValueSpec:
				collectCommentGroup(x.Doc, false, x.Pos())
				collectCommentGroup(x.Comment, true, x.Pos())
			case *ast.ImportSpec:
				collectCommentGroup(x.Doc, false, x.Pos())
				collectCommentGroup(x.Comment, true, x.Pos())
			case *ast.TypeSpec:
				collectCommentGroup(x.Doc, false, x.Pos())
				collectCommentGroup(x.Comment, true, x.Pos())
			case *ast.Field:
				collectCommentGroup(x.Doc, false, x.Pos())
				collectCommentGroup(x.Comment, true, x.Pos())
			}
			return true
		})
	}

	for i := range p.Package.Syntax {
		ast.Inspect(p.Package.Syntax[i], func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.CallExpr:
				fn := p.Package.TypesInfo.TypeOf(x.Fun)
				if fn != nil {
					if s, ok := fn.(*types.Signature); ok {
						if n, ok := p.signatures[s]; ok {
							switch n.(type) {
							case *ast.FuncDecl, *ast.FuncLit:
								// skip declared functions
							default:
								p.signatures[s] = x.Fun
							}
						} else {
							p.signatures[s] = x.Fun
						}
					}
				}
			}
			return true
		})
	}

	return p
}

type pkgInfo struct {
	u       *Universe
	Package *packages.Package

	imports map[string]Package

	constants map[string]*types.Const
	types     map[string]*types.TypeName
	funcs     map[string]*types.Func
	methods   map[*types.Named][]*types.Func

	endLineToCommentGroup         map[fileLine]*ast.CommentGroup
	endLineToTrailingCommentGroup map[fileLine]*ast.CommentGroup

	funcDecls  map[*types.Func]ast.Node
	signatures map[*types.Signature]ast.Node

	funcResultResolvers  sync.Map
	funcResultsResolving sync.Map

	sourceDir *string
}

func (pi *pkgInfo) SourceDir() string {
	if pi.sourceDir != nil {
		return *pi.sourceDir
	}

	sourceDir := func() string {
		if pi == nil || pi.Package == nil {
			return ""
		}
		if pi.Module() == nil {
			return ""
		}
		if pi.Package.PkgPath == pi.Module().Path {
			return pi.Module().Dir
		}
		return filepath.Join(pi.Module().Dir, pi.Package.PkgPath[len(pi.Module().Path):])
	}()

	pi.sourceDir = ptr.Ptr(sourceDir)

	return sourceDir
}

func (pi *pkgInfo) FileSet() *token.FileSet {
	return pi.u.fset
}

func (pi *pkgInfo) Pkg() *types.Package {
	return pi.Package.Types
}

func (pi *pkgInfo) ObjectOf(id *ast.Ident) types.Object {
	return pi.Package.TypesInfo.ObjectOf(id)
}

func (pi *pkgInfo) Module() *packages.Module {
	return pi.Package.Module
}

func (pi *pkgInfo) Imports() map[string]Package {
	return pi.imports
}

func (pi *pkgInfo) Files() []*ast.File {
	return pi.Package.Syntax
}

func (pi *pkgInfo) Eval(expr ast.Expr) (types.TypeAndValue, error) {
	return types.Eval(pi.Package.Fset, pi.Package.Types, expr.Pos(), StringifyNode(pi.Package.Fset, expr))
}

func (pi *pkgInfo) Constant(n string) *types.Const {
	return pi.constants[n]
}

func (pi *pkgInfo) Constants() map[string]*types.Const {
	return pi.constants
}

func (pi *pkgInfo) Type(n string) *types.TypeName {
	return pi.types[n]
}

func (pi *pkgInfo) Types() map[string]*types.TypeName {
	return pi.types
}

func (pi *pkgInfo) Function(n string) *types.Func {
	return pi.funcs[n]
}

func (pi *pkgInfo) Functions() map[string]*types.Func {
	return pi.funcs
}

func (pi *pkgInfo) MethodsOf(n *types.Named, ptr bool) []*types.Func {
	funcs, _ := pi.methods[n]

	if ptr {
		return funcs
	}

	notPtrMethods := make([]*types.Func, 0)

	for i := range funcs {
		s := funcs[i].Type().(*types.Signature)

		if _, ok := s.Recv().Type().(*types.Pointer); !ok {
			notPtrMethods = append(notPtrMethods, funcs[i])
		}
	}

	return notPtrMethods
}

func (pi *pkgInfo) Position(pos token.Pos) token.Position {
	return pi.Package.Fset.Position(pos)
}

func (pi *pkgInfo) Decl(pos token.Pos) ast.Decl {
	if f := pi.File(pos); f != nil {
		for _, d := range f.Decls {
			if d.Pos() <= pos && pos <= d.End() {
				return d
			}
		}
	}
	return nil
}

func (pi *pkgInfo) File(pos token.Pos) *ast.File {
	for _, f := range pi.Files() {
		if f.Pos() <= pos && pos <= f.End() {
			return f
		}
	}
	return nil
}

func (pi *pkgInfo) Doc(pos token.Pos) (map[string][]string, []string) {
	return ExtractCommentTags(commentLinesFrom(pi.priorCommentLines(pos, -1)))
}

func (pi *pkgInfo) Comment(pos token.Pos) []string {
	return commentLinesFrom(pi.priorCommentLines(pos, 0))
}

func (pi *pkgInfo) priorCommentLines(pos token.Pos, deltaLines int) *ast.CommentGroup {
	position := pi.Package.Fset.Position(pos)
	key := fileLine{position.Filename, position.Line + deltaLines}
	if deltaLines == 0 {
		// should ignore trailing comments
		// when deltaLines eq 0 means find trailing comments
		if lines, ok := pi.endLineToTrailingCommentGroup[key]; ok {
			return lines
		}
	}
	return pi.endLineToCommentGroup[key]
}

type fileLine struct {
	file string
	line int
}

func commentLinesFrom(commentGroups ...*ast.CommentGroup) (comments []string) {
	if len(commentGroups) == 0 {
		return nil
	}

	for _, commentGroup := range commentGroups {
		if commentGroup == nil {
			continue
		}

		for _, line := range strings.Split(strings.TrimSpace(commentGroup.Text()), "\n") {
			// skip go: prefix
			if strings.HasPrefix(line, "go:") {
				continue
			}
			comments = append(comments, line)
		}
	}
	return comments
}
