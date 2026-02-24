package inflector

import (
	"github.com/octohelm/gengo/pkg/inflector/internal"
)

// Pluralize 使用默认规则集把 s 转成复数形式。
func Pluralize(s string) string {
	return internal.Defaults.Inflected(internal.Plural, s)
}

// Singularize 使用默认规则集把 s 转成单数形式。
func Singularize(s string) string {
	return internal.Defaults.Inflected(internal.Singular, s)
}
