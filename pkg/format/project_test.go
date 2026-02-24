package format_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/format"
)

func TestProjectRunWithDeepEntrypoint(t *testing.T) {
	t.Run("GIVEN 使用 ./... 作为入口", func(t *testing.T) {
		tmpDir := t.TempDir()

		Must(t, func() error {
			files := map[string]string{
				"go.mod":                "module example.com/project\n\ngo 1.26.0\n",
				"root.go":               "package sample\n\nfunc Root(){\n}\n",
				"nested/child/child.go": "package child\n\nfunc Child(){\n}\n",
				"testdata/ignored.go":   "package ignored\n\nfunc Ignored(){\n}\n",
			}

			for name, content := range files {
				filename := filepath.Join(tmpDir, filepath.FromSlash(name))

				if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
					return err
				}

				if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
					return err
				}
			}

			return nil
		})

		wd := MustValue(t, os.Getwd)
		Must(t, func() error {
			return os.Chdir(tmpDir)
		})
		defer func() {
			Must(t, func() error {
				return os.Chdir(wd)
			})
		}()

		p := &format.Project{
			Entrypoint: []string{"./..."},
			Write:      true,
		}

		t.Run("WHEN 执行项目格式化", func(t *testing.T) {
			Must(t, func() error {
				return p.Init(context.Background())
			})
			Must(t, func() error {
				return p.Run(context.Background())
			})

			t.Run("THEN 根目录文件会被格式化", func(t *testing.T) {
				data := MustValue(t, func() ([]byte, error) {
					return os.ReadFile(filepath.Join(tmpDir, "root.go"))
				})

				Then(t, "根目录函数声明应带空格",
					Expect(strings.Contains(string(data), "func Root() {\n"), Be(cmp.True())),
				)
			})

			t.Run("THEN 深层子目录文件会被格式化", func(t *testing.T) {
				data := MustValue(t, func() ([]byte, error) {
					return os.ReadFile(filepath.Join(tmpDir, "nested", "child", "child.go"))
				})

				Then(t, "子目录函数声明应带空格",
					Expect(strings.Contains(string(data), "func Child() {\n"), Be(cmp.True())),
				)
			})

			t.Run("THEN testdata 目录仍然会被忽略", func(t *testing.T) {
				data := MustValue(t, func() ([]byte, error) {
					return os.ReadFile(filepath.Join(tmpDir, "testdata", "ignored.go"))
				})

				Then(t, "testdata 下文件不应被格式化",
					Expect(strings.Contains(string(data), "func Ignored(){\n"), Be(cmp.True())),
				)
			})
		})
	})
}
