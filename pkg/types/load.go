package types

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/packages"
)

const (
	LoadFiles     = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles
	LoadImports   = LoadFiles | packages.NeedImports
	LoadTypes     = LoadImports | packages.NeedTypes | packages.NeedTypesSizes
	LoadSyntax    = LoadTypes | packages.NeedSyntax | packages.NeedTypesInfo
	LoadAllSyntax = LoadSyntax | packages.NeedDeps | packages.NeedModule
)

func Load(patterns []string) (Universe, map[string]bool, error) {
	c := &packages.Config{
		Mode: LoadAllSyntax,
	}

	pkgs, err := packages.Load(c, patterns...)
	if err != nil {
		return nil, nil, err
	}

	u := Universe{}
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

			if _, ok := u[importedPkg.PkgPath]; !ok {
				register(importedPkg)
			}
		}

		u[p.PkgPath] = pkg

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

type Universe map[string]Package

func (u Universe) Package(pkgPath string) Package {
	return u[pkgPath]
}

func (u Universe) LocateInPackage(pos token.Pos) Package {
	for i := range u {
		p := u[i]
		for _, f := range p.Files() {
			if f.Pos() >= pos && pos <= f.End() {
				return p
			}
		}
	}
	return nil
}
