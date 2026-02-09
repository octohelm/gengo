package types

import (
	"errors"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestExtractCommentTags(t *testing.T) {
	t.Run("应该正确提取注释标签", func(t *testing.T) {
		commentLines := []string{
			"Human comment that is ignored.",
			"\tCode",
			"+gengo:test=value1",
			"@bar",
			"+baz=qux,zrb=true",
			"+gengo:test value2",
		}

		tags, others := ExtractCommentTags(commentLines, '+', '@')

		t.Run("THEN 标签应该正确提取", func(t *testing.T) {
			Then(t, "gengo:test 标签应该有两个值",
				Expect(tags["gengo:test"],
					Equal([]string{"value1", "value2"}),
				),
			)

			Then(t, "bar 标签应该有空值",
				Expect(tags["bar"],
					Equal([]string{""}),
				),
			)

			Then(t, "baz标签应该有复合值",
				Expect(tags["baz"],
					Equal([]string{"qux,zrb=true"}),
				),
			)

			Then(t, "所有标签应该匹配预期",
				Expect(tags,
					Equal(map[string][]string{
						"gengo:test": {"value1", "value2"},
						"bar":        {""},
						"baz":        {"qux,zrb=true"},
					}),
				),
			)
		})

		t.Run("THEN 非标签行应该正确分离", func(t *testing.T) {
			Then(t, "others 应该包含非标签行",
				Expect(others,
					Equal([]string{
						"Human comment that is ignored.",
						"\tCode",
					}),
				),
			)

			Then(t, "others 不应该包含标签行",
				Expect(others,
					Be(func(actual []string) error {
						for _, line := range actual {
							if len(line) > 0 && (line[0] == '+' || line[0] == '@') {
								return errors.New("包含标签行")
							}
						}
						return nil
					}),
				),
			)
		})
	})

	t.Run("边界条件测试", func(t *testing.T) {
		t.Run("空输入应该返回空结果", func(t *testing.T) {
			commentLines := make([]string, 0)

			tags, others := ExtractCommentTags(commentLines, '+', '@')

			Then(t, "标签map应该为空",
				Expect(len(tags), Equal(0)),
			)

			Then(t, "others应该为空切片",
				Expect(len(others), Equal(0)),
			)
		})

		t.Run("只包含普通注释应该正确分离", func(t *testing.T) {
			// GIVEN 只包含普通注释
			commentLines := []string{
				"Just a comment",
				"Another comment",
			}

			// WHEN 提取注释标签
			tags, others := ExtractCommentTags(commentLines, '+', '@')

			// THEN
			Then(t, "标签map应该为空",
				Expect(len(tags), Equal(0)),
			)

			Then(t, "others应该包含所有行",
				Expect(others,
					Equal([]string{"Just a comment", "Another comment"}),
				),
			)
		})

		t.Run("混合标签和注释应该正确处理", func(t *testing.T) {
			commentLines := []string{
				"Start",
				"+tag1",
				"Middle",
				"@tag2=value",
				"End",
			}

			tags, others := ExtractCommentTags(commentLines, '+', '@')

			Then(t, "应该提取两个标签",
				Expect(len(tags), Equal(2)),
			)

			Then(t, "标签值应该正确",
				Expect(tags,
					Equal(map[string][]string{
						"tag1": {""},
						"tag2": {"value"},
					}),
				),
			)

			Then(t, "others应该包含三行非标签",
				Expect(others,
					Equal([]string{"Start", "Middle", "End"}),
				),
			)
		})
	})

	t.Run("特殊字符处理", func(t *testing.T) {
		t.Run("重复标签应该合并值", func(t *testing.T) {
			commentLines := []string{
				"+same=first",
				"+same=second",
				"+same third",
			}

			tags, _ := ExtractCommentTags(commentLines, '+', '@')

			Then(t, "重复标签的值应该合并",
				Expect(tags["same"],
					Equal([]string{"first", "second", "third"}),
				),
			)
		})
	})

	t.Run("错误处理场景", func(t *testing.T) {
		t.Run("nil 输入应该安全处理", func(t *testing.T) {
			Then(t, "nil输入不应该panic",
				ExpectMust(func() error {
					_, _ = ExtractCommentTags(nil, '+', '@')
					return nil
				}),
			)
		})

		t.Run("包含空格的标签应该正确解析", func(t *testing.T) {
			commentLines := []string{
				"+tag=value with spaces",
				"+another=more spaces here",
			}

			tags, _ := ExtractCommentTags(commentLines, '+', '@')

			Then(t, "包含空格的标签值应该完整保留",
				Expect(tags["tag"],
					Equal([]string{"value with spaces"}),
				),
			)

			Then(t, "另一个标签值也应该正确",
				Expect(tags["another"],
					Equal([]string{"more spaces here"}),
				),
			)
		})
	})
}
