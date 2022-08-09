package types

import (
	"bytes"
	"strings"
)

// ExtractCommentTags parses comments for lines of the form:
//
//	'marker' + "key=value".
//
// Values are optional; "" is the default.  A tag can be specified more than
// one time and all values are returned.  If the resulting map has an entry for
// a key, the value (a slice) is guaranteed to have at least 1 element.
//
// Example: if you pass '+' or '@' for 'marker', and the following lines are in
// the comments:
//
//	+foo=value1
//	+bar
//	+foo value2
//	+baz="qux"
//
// Then this function will return:
//
//	map[string][]string{"foo":{"value1, "value2"}, "bar": {"true"}, "baz": {"qux"}}
func ExtractCommentTags(lines []string, markers ...byte) (tags map[string][]string, otherLines []string) {
	if len(markers) == 0 {
		markers = []byte{'+', '@'}
	}

	tags = map[string][]string{}

	for _, line := range lines {
		line = strings.Trim(line, " ")

		if !(len(line) != 0 && oneOf(markers, line[0])) {
			otherLines = append(otherLines, line)
			continue
		}

		k, v := splitKV(line[1:])
		tags[k] = append(tags[k], v)
	}

	return
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

func oneOf(markers []byte, b byte) bool {
	for _, m := range markers {
		if b == m {
			return true
		}
	}
	return false
}
