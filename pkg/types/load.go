package types

import (
	"fmt"
	"go/token"
	"iter"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

const (
	LoadFiles     = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles
	LoadImports   = LoadFiles | packages.NeedImports
	LoadTypes     = LoadImports | packages.NeedTypes | packages.NeedTypesSizes
	LoadSyntax    = LoadTypes | packages.NeedSyntax | packages.NeedTypesInfo
	LoadAllSyntax = LoadSyntax | packages.NeedDeps | packages.NeedModule
)

func Load(patterns []string) (*Universe, error) {
	fset := token.NewFileSet()

	c := &packages.Config{
		Fset: fset,
		Mode: LoadAllSyntax,
	}

	pkgs, err := packages.Load(c, patterns...)
	if err != nil {
		return nil, err
	}

	u := &Universe{
		fset: fset,
		pkgs: map[string]Package{},
	}

	directPkgPaths := map[string]bool{}
	localPkgPaths := map[string]bool{}
	rootPkgPaths := map[string]bool{}

	var register func(p *packages.Package)
	register = func(p *packages.Package) {
		if len(p.Errors) > 0 {
			for i := range p.Errors {
				e := p.Errors[i]
				fmt.Println("[warning]", e.Pos, e.Msg)
			}
		}

		pkg := newPkg(p, u)

		for k := range p.Imports {
			importedPkg := p.Imports[k]

			if _, ok := u.pkgs[importedPkg.PkgPath]; !ok {
				register(importedPkg)
			}
		}

		u.pkgs[p.PkgPath] = pkg

		for rootPkgPath := range rootPkgPaths {
			// when is sub pkg of root pkg
			if p.Module != nil && rootPkgPath == p.Module.Path {
				localPkgPaths[p.PkgPath] = directPkgPaths[p.PkgPath]
			}
		}
	}

	for i := range pkgs {
		p := pkgs[i]
		if len(p.Errors) > 0 {
			for i := range p.Errors {
				e := p.Errors[i]
				if e.Kind == packages.ListError {
					return nil, e
				}
			}
		}
		rootPkgPaths[p.Module.Path] = true
		directPkgPaths[p.PkgPath] = true
	}

	for i := range pkgs {
		register(pkgs[i])
	}

	u.localPkgPaths = localPkgPaths

	return u, nil
}

type Universe struct {
	fset          *token.FileSet
	pkgs          map[string]Package
	localPkgPaths map[string]bool
}

func (v *Universe) LocalPkgPaths() iter.Seq2[string, bool] {
	return func(yield func(string, bool) bool) {
		for pkgPath, direct := range v.localPkgPaths {
			if !yield(pkgPath, direct) {
				return
			}
		}
	}
}

func (u *Universe) Package(pkgPath string) Package {
	v, _ := u.pkgs[pkgPath]
	return v
}

func (u *Universe) LocateInPackage(pos token.Pos) Package {
	pp := u.fset.Position(pos)
	dir := filepath.Dir(pp.Filename)

	for _, p := range u.pkgs {
		if dir == p.SourceDir() {
			return p
		}
	}
	return nil
}
