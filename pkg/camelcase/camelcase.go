package camelcase

import (
	"unicode"
	"unicode/utf8"
)

// Split 会把 CamelCase 风格的字符串拆成单词切片。
//
// 它同时支持数字，并兼容 lower camel case 和 Upper CamelCase。
//
// 拆分规则：
//
//  1. 如果字符串不是合法 UTF-8，则直接返回仅包含原字符串的切片。
//  2. 将所有 Unicode 字符归类为四组：小写字母、大写字母、数字和其他字符。
//  3. 遍历字符串，并在相邻字符分类变化时引入拆分点。
//  4. 对连续大写后跟小写的场景做一次调整，把大写串末尾字符移到后一个单词前面。
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

	return entries
}

// RuneType 表示标识符拆分过程中 rune 的分类。
type RuneType int

const (
	// RuneOther 表示既不是字母也不是数字的 rune。
	RuneOther = iota
	// RuneLower 表示小写字母。
	RuneLower
	// RuneUpper 表示大写字母。
	RuneUpper
	// RuneDigit 表示十进制数字。
	RuneDigit
)
