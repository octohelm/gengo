package gengo

import (
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
)

// 这些是生成器常用的命名辅助函数重导出。
var (
	UpperSnakeCase = camelcase.UpperSnakeCase
	LowerSnakeCase = camelcase.LowerSnakeCase
	UpperKebabCase = camelcase.UpperKebabCase
	LowerKebabCase = camelcase.LowerKebabCase
	UpperCamelCase = camelcase.UpperCamelCase
	LowerCamelCase = camelcase.LowerCamelCase
)

// ImportGoPath 会在 import path 中存在 vendor 前缀时将其去掉。
func ImportGoPath(importPath string) string {
	i := strings.LastIndex(importPath, "/vendor/")
	if i > 0 {
		return importPath[i:]
	}
	return importPath
}

// PkgImportPathAndExpose 将限定标识符拆成包路径和导出名。
func PkgImportPathAndExpose(s string) (string, string) {
	if i := strings.Index(s, "["); i > 0 {
		s = s[0:i]
	}
	if i := strings.LastIndex(s, "."); i > 0 {
		return ImportGoPath(s[0:i]), s[i+1:]
	}
	return "", s
}
