package types

import (
	"bytes"
	"slices"
	"strings"
)

// ExtractCommentTags 解析类似指令的注释行，并返回标签以及非标签内容。
//
// 它识别如下格式的行：
//
//	'marker' + "key=value".
//
// 值是可选的，默认值为空字符串。一个标签可以重复出现多次，返回时会合并全部值。
// 只要结果 map 中存在某个 key，对应的 value 切片就至少会有一个元素。
//
// 例如，当 marker 传入 '+' 或 '@'，并且注释中包含以下内容：
//
//	+foo=value1
//	+bar
//	+foo value2
//	+baz="qux"
//
// 那么这个函数会返回：
//
//	map[string][]string{"foo": {"value1", "value2"}, "bar": {""}, "baz": {"qux"}}
func ExtractCommentTags(lines []string, markers ...byte) (tags map[string][]string, otherLines []string) {
	if len(markers) == 0 {
		markers = []byte{'+', '@'}
	}

	tags = map[string][]string{}

	for _, line := range lines {
		line = strings.Trim(line, " ")

		if !(len(line) != 0 && slices.Contains(markers, line[0])) {
			otherLines = append(otherLines, line)
			continue
		}

		k, v := splitKV(line[1:])
		tags[k] = append(tags[k], v)
	}

	return tags, otherLines
}

func splitKV(line string) (string, string) {
	k := bytes.NewBuffer(nil)
	v := bytes.NewBuffer(nil)

	forValue := false

	for _, c := range line {
		if !forValue && (c == '=' || c == ' ') {
			forValue = true
			continue
		}

		if forValue {
			v.WriteRune(c)
		} else {
			k.WriteRune(c)
		}
	}

	return k.String(), v.String()
}
