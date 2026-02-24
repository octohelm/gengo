package camelcase

import (
	"bytes"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// 这些是基于 Split 构建的常用大小写转换函数。
var (
	// LowerSnakeCase 将名称转换为 lower snake_case。
	LowerSnakeCase = makeCase("_", wrap(strings.ToLower))
	// UpperSnakeCase 将名称转换为 upper SNAKE_CASE。
	UpperSnakeCase = makeCase("_", wrap(strings.ToUpper))
	// LowerKebabCase 将名称转换为 lower kebab-case。
	LowerKebabCase = makeCase("-", wrap(strings.ToLower))
	// UpperKebabCase 将名称转换为 upper KEBAB-CASE。
	UpperKebabCase = makeCase("-", wrap(strings.ToUpper))
	// LowerCamelCase 将名称转换为 lowerCamelCase。
	LowerCamelCase = makeCase("", func(w string, i int) string {
		if i == 0 {
			return strings.ToLower(w)
		}
		if bytes.EqualFold([]byte(w), []byte("ID")) {
			return "ID"
		}
		return cases.Title(language.Und).String(w)
	})
	// UpperCamelCase 将名称转换为 UpperCamelCase。
	UpperCamelCase = makeCase("", func(w string, i int) string {
		if bytes.EqualFold([]byte(w), []byte("ID")) {
			return "ID"
		}
		return cases.Title(language.Und).String(w)
	})
)

func wrap(transWord func(w string) string) func(w string, i int) string {
	return func(w string, i int) string {
		return transWord(w)
	}
}

func makeCase(linker string, transWord func(w string, i int) string) func(s string) string {
	return func(s string) string {
		words := Split(s)

		var b strings.Builder
		idx := 0

		for _, word := range words {
			if len(word) == 1 && unicode.IsGraphic(rune(word[0])) {
				c := rune(word[0])

				if !(unicode.IsDigit(c) || unicode.IsLetter(c)) {
					// only number or letter
					continue
				}
			}
			if idx > 0 {
				b.WriteString(linker)
			}
			b.WriteString(transWord(word, idx))
			idx++
		}

		return b.String()
	}
}
