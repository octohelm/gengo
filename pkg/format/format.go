package format

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/ast/astutil"
	"mvdan.cc/gofumpt/format"
)

func OptionsFromModFile(file *modfile.File) Options {
	o := Options{}
	o.LocalGroupPrefix = file.Module.Mod.Path

	for _, stmt := range file.Syntax.Stmt {
		switch x := stmt.(type) {
		case *modfile.LineBlock:
			for _, line := range x.Line {
				line.Comments.Before = slices.Concat(x.Before, line.Comments.Before)
			}
		}
	}

	for _, r := range file.Require {
		if ig, ok := importGroup(r.Syntax.Comments.Before); ok {
			if o.ImportGroups == nil {
				o.ImportGroups = map[string]*ImportGroup{}
			}
			i, ok := o.ImportGroups[ig]
			if !ok {
				i = &ImportGroup{}
				o.ImportGroups[ig] = i
			}
			i.Prefixes = append(i.Prefixes, r.Mod.Path)
		}
	}

	return o
}

const importGroupDirective = "+gengo:import:group="

func importGroup(comments []modfile.Comment) (string, bool) {
	for _, comment := range comments {
		if i := strings.Index(comment.Token, importGroupDirective); i > 0 {
			return strings.TrimSpace(comment.Token[i+len(importGroupDirective):]), true
		}
	}
	return "", false
}

type Options struct {
	LocalGroupPrefix string
	ImportGroups     map[string]*ImportGroup
}

type ImportGroup struct {
	Prefixes []string
}

func (g *ImportGroup) Match(path string) bool {
	for _, p := range g.Prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func Source(src []byte, opt Options) ([]byte, error) {
	fset := token.NewFileSet()
	fset.AddFile("gengo_base.fake.go", 1, 10)

	file, err := parser.ParseFile(fset, "", src, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	is := &importSorter{
		LocalPkgPathPrefix: opt.LocalGroupPrefix,
		ImportGroups:       opt.ImportGroups,
	}

	src, err = is.File(fset, file)
	if err != nil {
		return nil, err
	}

	return format.Source(src, format.Options{})
}

type importSorter struct {
	LocalPkgPathPrefix string
	ImportGroups       map[string]*ImportGroup

	importSpecs     []*ast.ImportSpec
	sideImportSpecs []*ast.ImportSpec
}

func (s *importSorter) File(fset *token.FileSet, file *ast.File) ([]byte, error) {
	s.collectAndRemove(file)

	importDecls := make([]ast.Decl, 0, 3)

	if specGroups := s.groupAndSort(s.importSpecs); len(specGroups) > 0 {
		if decl := s.declFromImportSpecGroups(specGroups...); decl != nil {
			importDecls = append(importDecls, decl)
		}
	}

	if specGroups := s.groupAndSort(s.sideImportSpecs); len(specGroups) > 0 {
		if decl := s.declFromImportSpecGroups(specGroups...); decl != nil {
			importDecls = append(importDecls, decl)
		}
	}

	file.Decls = slices.Concat(importDecls, file.Decls)

	buf := bytes.NewBuffer(nil)
	if err := printer.Fprint(buf, fset, file); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *importSorter) isImportC(importSpec *ast.ImportSpec) bool {
	return importSpec.Path.Value == `"C"`
}

func (s *importSorter) collectAndRemove(file *ast.File) {
	for _, importSpec := range file.Imports {
		if s.isImportC(importSpec) {
			continue
		}

		if importSpec.Name != nil && importSpec.Name.Name == "_" {
			s.sideImportSpecs = append(s.sideImportSpecs, importSpec)
			continue
		}

		s.importSpecs = append(s.importSpecs, importSpec)
	}

	astutil.Apply(file, func(cursor *astutil.Cursor) bool {
		if importSpec, ok := cursor.Node().(*ast.ImportSpec); ok {
			if s.isImportC(importSpec) {
				return false
			}
			cursor.Delete()
			return false
		}
		return true
	}, nil)
}

func (s *importSorter) declFromImportSpecGroups(specGroups ...[]*ast.ImportSpec) *ast.GenDecl {
	if len(specGroups) == 0 {
		return nil
	}

	grouped := len(specGroups) > 1 // only c imports without grouped

	decl := &ast.GenDecl{
		Tok:    token.IMPORT,
		Rparen: 1,
		Lparen: 1,
	}

	for i, specGroup := range specGroups {
		for j, spec := range specGroup {
			importSpec := &ast.ImportSpec{
				Path: &ast.BasicLit{
					Value: spec.Path.Value,
				},
			}

			if spec.Name != nil {
				importSpec.Name = &ast.Ident{
					Name: spec.Name.Name,
				}
			}

			if i > 0 && j == 0 {
				if grouped {
					if spec.Name != nil {
						importSpec.Name.Name = "\n\n" + importSpec.Name.Name
					} else {
						importSpec.Path.Value = "\n\n" + importSpec.Path.Value
					}
				}
			}

			decl.Specs = append(decl.Specs, importSpec)
		}
	}

	if len(decl.Specs) == 0 {
		return nil
	}

	return decl
}

func (s *importSorter) groupAndSort(specs []*ast.ImportSpec) [][]*ast.ImportSpec {
	if len(specs) == 0 {
		return nil
	}

	importSpecGroups := make([][]*ast.ImportSpec, 1+1+len(s.ImportGroups)+1)

	stdIndex := 0
	otherVendorIndex := 1
	groupIndexStart := 2
	localIndex := len(importSpecGroups) - 1

	for _, spec := range specs {
		pkgPath, _ := strconv.Unquote(spec.Path.Value)

		if s.LocalPkgPathPrefix != "" && strings.HasPrefix(pkgPath, s.LocalPkgPathPrefix) {
			importSpecGroups[localIndex] = append(importSpecGroups[localIndex], spec)
			continue
		}

		matched := false
		for i, g := range slices.Sorted(maps.Keys(s.ImportGroups)) {
			importGroup := s.ImportGroups[g]
			groupIndex := groupIndexStart + i

			if importGroup.Match(pkgPath) {
				importSpecGroups[groupIndex] = append(importSpecGroups[groupIndex], spec)
				matched = true
				break
			}
		}

		if matched {
			continue
		}

		if strings.Contains(pkgPath, ".") {
			importSpecGroups[otherVendorIndex] = append(importSpecGroups[otherVendorIndex], spec)
			continue
		}

		importSpecGroups[stdIndex] = append(importSpecGroups[stdIndex], spec)
	}

	for _, importSpecs := range importSpecGroups {
		sort.Slice(importSpecs, func(i, j int) bool { return importSpecs[i].Path.Value < importSpecs[j].Path.Value })
	}

	return importSpecGroups
}
