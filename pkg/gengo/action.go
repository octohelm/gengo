package gengo

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
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
// excludePrefix 用于排除生成文件（通常为 OutputFileBaseName）。
func computePkgContentHash(u *gengotypes.Universe, pkgPath string, excludePrefix string) string {
	self := depSummary(u.Package(pkgPath), excludePrefix)
	if self == "" {
		return ""
	}

	// 收集传递依赖摘要
	var summaries []string
	collectDeps(u, pkgPath, map[string]bool{pkgPath: true}, func(depPath string) {
		if s := depSummary(u.Package(depPath), excludePrefix); s != "" {
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

// depSummary 返回 pkg 的简单摘要：本地包用 dirhash（排除生成文件），外部依赖用 module@version。
func depSummary(pkg gengotypes.Package, excludePrefix string) string {
	if pkg == nil {
		return ""
	}
	mod := pkg.Module()
	if mod == nil {
		return ""
	}
	dir := pkg.SourceDir()

	// 同一模块下的包（SourceDir 以 mod.Dir 为前缀）→ dirhash，排除生成文件
	if dir != "" && strings.HasPrefix(dir, mod.Dir) {
		return hashDirExcluding(dir, excludePrefix)
	}

	// 外部依赖 → module@version
	return mod.Path + "@" + mod.Version
}

// hashDirExcluding 对目录计算 dirhash.Hash1 摘要，排除名称以 excludePrefix+"." 开头的文件。
func hashDirExcluding(dir string, excludePrefix string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if excludePrefix != "" && strings.HasPrefix(name, excludePrefix+".") {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)

	h, _ := dirhash.Hash1(names, func(name string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(dir, name))
	})
	return h
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
