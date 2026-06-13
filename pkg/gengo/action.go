package gengo

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"

	gengotypes "github.com/octohelm/gengo/pkg/types"
)

// FrameworkVersion 定义 gengo 框架版本，breaking change 时手动更新。
const FrameworkVersion = "v1"

// computeActionID 为 (生成器, 包) 组合计算 Layer 2 actionID。
func computeActionID(genName, genVersion, pkgContentHash string) string {
	h := sha256.New()
	h.Write([]byte(FrameworkVersion + "\n"))
	h.Write([]byte(genName + "\n"))
	h.Write([]byte(genVersion + "\n"))
	h.Write([]byte(pkgContentHash + "\n"))
	return hex.EncodeToString(h.Sum(nil))
}

// generatorVersion 从编译时 BuildInfo 中获取生成器所在模块的版本。
func generatorVersion(g Generator) string {
	t := reflect.TypeOf(g)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	pkgPath := t.PkgPath()

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}

	if info.Main.Path == pkgPath || strings.HasPrefix(pkgPath, info.Main.Path+"/") {
		return "(devel)"
	}

	for _, dep := range info.Deps {
		if dep.Path == pkgPath || strings.HasPrefix(pkgPath, dep.Path+"/") {
			if dep.Version != "" {
				return dep.Version
			}
			return "(devel)"
		}
	}

	return "(unknown)"
}

// computePkgContentHash 为 pkgPath 计算 Layer 1 内容摘要。
func computePkgContentHash(u *gengotypes.Universe, pkgPath string) string {
	self := depSummary(u.Package(pkgPath))
	if self == "" {
		return ""
	}

	// 收集传递依赖摘要
	var summaries []string
	collectDeps(u, pkgPath, map[string]bool{pkgPath: true}, func(depPath string) {
		if s := depSummary(u.Package(depPath)); s != "" {
			summaries = append(summaries, s)
		}
	})
	slices.Sort(summaries)

	h := sha256.New()
	h.Write([]byte(self + "\n"))
	for _, s := range summaries {
		h.Write([]byte(s + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// depSummary 返回 pkg 的简单摘要：本地包用 dirhash，外部依赖用 module@version。
func depSummary(pkg gengotypes.Package) string {
	if pkg == nil {
		return ""
	}
	mod := pkg.Module()
	if mod == nil {
		return ""
	}
	dir := pkg.SourceDir()

	// 同一模块下的包（SourceDir 以 mod.Dir 为前缀）→ dirhash
	if dir != "" && strings.HasPrefix(dir, mod.Dir) {
		h, err := dirhash.HashDir(dir, pkg.Pkg().Path(), dirhash.Hash1)
		if err != nil {
			return ""
		}
		return h
	}

	// 外部依赖 → module@version
	return mod.Path + "@" + mod.Version
}

// collectDeps 遍历 pkgPath 的传递依赖，对每个调用 fn。
func collectDeps(u *gengotypes.Universe, pkgPath string, seen map[string]bool, fn func(string)) {
	pkg := u.Package(pkgPath)
	if pkg == nil {
		return
	}
	for impPath := range pkg.Imports() {
		if seen[impPath] {
			continue
		}
		seen[impPath] = true
		fn(impPath)
		collectDeps(u, impPath, seen, fn)
	}
}
