package gengo

import (
	"strings"
	"unicode"

	"github.com/go-courier/gengo/pkg/camelcase"
)

func UpperSnakeCase(s string) string {
	return rewords(s, func(result string, word string, idx int) string {
		newWord := strings.ToUpper(word)
		if idx == 0 || (len(newWord) == 1 && unicode.IsDigit(rune(newWord[0]))) {
			return result + newWord
		}
		return result + "_" + newWord
	})
}

func LowerSnakeCase(s string) string {
	return rewords(s, func(result string, word string, idx int) string {
		newWord := strings.ToLower(word)
		if idx == 0 || (len(newWord) == 1 && unicode.IsDigit(rune(newWord[0]))) {
			return result + newWord
		}
		return result + "_" + newWord
	})
}

func UpperCamelCase(s string) string {
	return rewords(s, func(result string, word string, idx int) string {
		return result + strings.Title(word)
	})
}

func LowerCamelCase(s string) string {
	return rewords(s, func(result string, word string, idx int) string {
		if idx == 0 {
			return result + strings.ToLower(word)
		}
		return result + strings.Title(word)
	})
}

func rewords(s string, reducer func(result string, word string, index int) string) string {
	words := camelcase.Split(s)

	var result = ""

	for idx, word := range words {
		if len(word) == 1 && unicode.IsGraphic(rune(word[0])) {
			c := rune(word[0])

			if !(unicode.IsDigit(c) || unicode.IsLetter(c)) {
				// only number or letter
				continue
			}
		}

		result = reducer(result, word, idx)
	}

	return result
}
