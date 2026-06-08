package types

import (
	"fmt"
	"go/token"
	"iter"
	"maps"
	"path/filepath"
	"slices"

	"golang.org/x/tools/go/packages"
)

const (
	// LoadFiles 加载包名和文件列表。
	LoadFiles = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles
	// LoadImports 在 LoadFiles 基础上额外加载导入包信息。
	LoadImports = LoadFiles | packages.NeedImports
	// LoadTypes 在 LoadImports 基础上额外加载类型信息。
	LoadTypes = LoadImports | packages.NeedTypes | packages.NeedTypesSizes
	// LoadSyntax 在 LoadTypes 基础上额外加载语法树和类型细节。
	LoadSyntax = LoadTypes | packages.NeedSyntax | packages.NeedTypesInfo
	// LoadAllSyntax 在 LoadSyntax 基础上额外加载依赖和模块元信息。
	LoadAllSyntax = LoadSyntax | packages.NeedDeps | packages.NeedModule
	// LoadForHash 加载计算摘要所需的最小信息（不含类型与语法）。
	LoadForHash = LoadImports | packages.NeedDeps | packages.NeedModule
)

// WithDir 为 Load 设置 packages.Config.Dir。
func WithDir(dir string) func(c *packages.Config) {
	return func(c *packages.Config) {
		c.Dir = dir
	}
}

// WithMode 为 Load 设置 packages.Config.Mode，覆盖默认的 LoadAllSyntax。
func WithMode(mode packages.LoadMode) func(c *packages.Config) {
	return func(c *packages.Config) {
		c.Mode = mode
	}
}

// Load 使用 go/packages 把 patterns 解析成一个 Universe。
func Load(patterns []string, options ...func(c *packages.Config)) (*Universe, error) {
	fset := token.NewFileSet()

	c := &packages.Config{
		Fset: fset,
		Mode: LoadAllSyntax,
	}

	for _, opt := range options {
		opt(c)
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

// Universe 保存已加载的包图以及相关元信息。
type Universe struct {
	fset          *token.FileSet
	pkgs          map[string]Package
	localPkgPaths map[string]bool
}

// LocalPkgPaths 迭代本地包路径，并标记它是否是直接加载目标。
func (v *Universe) LocalPkgPaths() iter.Seq2[string, bool] {
	return func(yield func(string, bool) bool) {
		for _, pkgPath := range slices.Sorted(maps.Keys(v.localPkgPaths)) {
			if !yield(pkgPath, v.localPkgPaths[pkgPath]) {
				return
			}
		}
	}
}

// Package 返回 pkgPath 对应的已加载包。
func (u *Universe) Package(pkgPath string) Package {
	v, _ := u.pkgs[pkgPath]
	return v
}

// LocateInPackage 返回拥有 pos 的包。
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
