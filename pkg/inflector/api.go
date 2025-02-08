package inflector

import "github.com/octohelm/gengo/pkg/inflector/internal"

func Pluralize(s string) string {
	return internal.Defaults.Inflected(internal.Plural, s)
}

func Singularize(s string) string {
	return internal.Defaults.Inflected(internal.Singular, s)
}
