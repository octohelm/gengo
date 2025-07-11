package camelcase

import (
	"unicode"
	"unicode/utf8"
)

// Split splits the camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
// Splitting rules
//
//  1. If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2. Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3. Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4. Iterate through array of split strings, and if a given string
//     is upper case:
//     if subsequent string is lower case:
//     move last character of upper case string to beginning of
//     lower case string
func Split(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}

	entries = make([]string, 0)

	runes := make([][]rune, 0, len(src))
	lastClass := 0
	class := 0

	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = RuneLower
		case unicode.IsUpper(r):
			class = RuneUpper
		case unicode.IsDigit(r):
			class = RuneDigit
		default:
			class = RuneOther
		}

		if class == lastClass || (class == RuneDigit && (lastClass == RuneUpper || lastClass == RuneLower)) || (class == RuneLower && lastClass == RuneDigit) {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
			lastClass = class
			continue
		}

		runes = append(runes, []rune{r})
		lastClass = class
	}

	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}

	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}

	return
}

type RuneType int

const (
	RuneOther = iota
	RuneLower
	RuneUpper
	RuneDigit
)
