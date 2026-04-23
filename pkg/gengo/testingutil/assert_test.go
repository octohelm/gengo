package testingutil

import (
	"regexp"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"
)

func TestFileChecker(t *testing.T) {
	files := map[string]string{
		"sample/zz_generated.demo.go": "package sample\n\nfunc runtimeDoc() {}\nfunc Demo() {}\n",
	}

	Then(t, "文件存在且内容满足组合谓词",
		Expect(files, Be(File("sample/zz_generated.demo.go",
			Contains("package sample", "func Demo()"),
			NotContains("func Missing()"),
			Count("func ", cmp.Eq(2)),
		))),
	)

	Then(t, "缺失文件应返回错误",
		Expect(File("missing.go")(files), ErrorMatch(regexp.MustCompile("生成文件不存在"))),
	)

	Then(t, "缺失片段应返回错误",
		Expect(Contains("missing")(files["sample/zz_generated.demo.go"]), ErrorMatch(regexp.MustCompile("内容缺少片段"))),
	)

	Then(t, "禁止片段存在时应返回错误",
		Expect(NotContains("func Demo()")(files["sample/zz_generated.demo.go"]), ErrorMatch(regexp.MustCompile("内容不应包含片段"))),
	)

	Then(t, "片段次数不满足时应返回错误",
		Expect(Count("func ", cmp.Eq(1))(files["sample/zz_generated.demo.go"]), ErrorMatch(regexp.MustCompile("出现次数不满足预期"))),
	)
}
