package gengo

import (
	"github.com/octohelm/gengo/pkg/camelcase"
	"strings"
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
	parts := strings.Split(importPath, "/vendor/")
	return parts[len(parts)-1]
}

func PkgImportPathAndExpose(s string) (string, string) {
	args := strings.Split(s, ".")
	lenOfArgs := len(args)
	if lenOfArgs > 1 {
		return ImportGoPath(strings.Join(args[0:lenOfArgs-1], ".")), args[lenOfArgs-1]
	}
	return "", s
}
