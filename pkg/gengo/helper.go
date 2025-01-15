package gengo

import (
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
)

var (
	UpperSnakeCase = camelcase.UpperSnakeCase
	LowerSnakeCase = camelcase.LowerSnakeCase
	UpperKebabCase = camelcase.UpperKebabCase
	LowerKebabCase = camelcase.LowerKebabCase
	UpperCamelCase = camelcase.UpperCamelCase
	LowerCamelCase = camelcase.LowerCamelCase
)

func ImportGoPath(importPath string) string {
	i := strings.LastIndex(importPath, "/vendor/")
	if i > 0 {
		return importPath[i:]
	}
	return importPath
}

func PkgImportPathAndExpose(s string) (string, string) {
	if i := strings.Index(s, "["); i > 0 {
		s = s[0:i]
	}
	if i := strings.LastIndex(s, "."); i > 0 {
		return ImportGoPath(s[0:i]), s[i+1:]
	}
	return "", s
}
