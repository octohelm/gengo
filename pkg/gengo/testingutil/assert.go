package testingutil

import (
	"fmt"
	"strings"
)

// File 返回要求生成文件存在且内容满足 checkers 的谓词。
func File(name string, checkers ...func(string) error) func(map[string]string) error {
	return func(files map[string]string) error {
		content, ok := files[name]
		if !ok {
			return fmt.Errorf("生成文件不存在: %s", name)
		}

		for _, check := range checkers {
			if err := check(content); err != nil {
				return err
			}
		}

		return nil
	}
}

// Contains 返回要求字符串包含所有片段的谓词。
func Contains(fragments ...string) func(string) error {
	return func(content string) error {
		for _, fragment := range fragments {
			if !strings.Contains(content, fragment) {
				return fmt.Errorf("内容缺少片段 %q\n%s", fragment, content)
			}
		}
		return nil
	}
}

// NotContains 返回要求字符串不包含任何片段的谓词。
func NotContains(fragments ...string) func(string) error {
	return func(content string) error {
		for _, fragment := range fragments {
			if strings.Contains(content, fragment) {
				return fmt.Errorf("内容不应包含片段 %q\n%s", fragment, content)
			}
		}
		return nil
	}
}

// Count 返回要求字符串中片段出现次数满足 check 的谓词。
func Count(fragment string, check func(int) error) func(string) error {
	return func(content string) error {
		count := strings.Count(content, fragment)
		if err := check(count); err != nil {
			return fmt.Errorf("片段 %q 出现次数不满足预期: %w\n%s", fragment, err, content)
		}
		return nil
	}
}
