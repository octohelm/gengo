package types

import (
	"fmt"
	"go/token"
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

func Load(patterns []string) (*Universe, map[string]bool, error) {
	fset := token.NewFileSet()

	c := &packages.Config{
		Fset: fset,
		Mode: LoadAllSyntax,
	}

	pkgs, err := packages.Load(c, patterns...)
	if err != nil {
		return nil, nil, err
	}

	u := &Universe{
		fset: fset,
		pkgs: map[string]Package{},
	}
	rootPkgPaths := map[string]bool{}
	pkgPaths := map[string]bool{}

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
				pkgPaths[p.PkgPath] = true
			}
		}
	}

	for i := range pkgs {
		p := pkgs[i]

		if len(p.Errors) > 0 {
			for i := range p.Errors {
				e := p.Errors[i]
				if e.Kind == packages.ListError {
					return nil, nil, e
				}
			}
		}

		rootPkgPaths[p.Module.Path] = true
	}

	for i := range pkgs {
		register(pkgs[i])
	}

	return u, pkgPaths, nil
}

type Universe struct {
	fset *token.FileSet
	pkgs map[string]Package
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
