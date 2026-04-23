package format_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
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

func TestProjectRejectEntrypointOutsideRoot(t *testing.T) {
	t.Run("GIVEN 入口指向当前项目根目录外", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")
		outsideDir := filepath.Join(tmpDir, "outside")

		Must(t, func() error {
			files := map[string]string{
				filepath.Join(projectDir, "go.mod"):     "module example.com/project\n\ngo 1.26.0\n",
				filepath.Join(projectDir, "root.go"):    "package sample\n\nfunc Root(){\n}\n",
				filepath.Join(outsideDir, "outside.go"): "package outside\n\nfunc Outside(){\n}\n",
			}

			for name, content := range files {
				if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
					return err
				}
			}

			return nil
		})

		wd := MustValue(t, os.Getwd)
		Must(t, func() error {
			return os.Chdir(projectDir)
		})
		defer func() {
			Must(t, func() error {
				return os.Chdir(wd)
			})
		}()

		p := &format.Project{
			Entrypoint: []string{filepath.Join("..", "outside", "outside.go")},
			Write:      true,
		}

		t.Run("WHEN 执行项目格式化", func(t *testing.T) {
			Must(t, func() error {
				return p.Init(context.Background())
			})

			Then(t, "THEN 应拒绝越出项目根目录",
				ExpectDo(func() error {
					return p.Run(context.Background())
				}, ErrorMatch(regexp.MustCompile("超出项目根目录"))),
			)

			data := MustValue(t, func() ([]byte, error) {
				return os.ReadFile(filepath.Join(outsideDir, "outside.go"))
			})

			Then(t, "THEN 项目外文件不应被格式化",
				Expect(strings.Contains(string(data), "func Outside(){\n"), Be(cmp.True())),
			)
		})
	})
}

func TestProjectRejectSymlinkOutsideRoot(t *testing.T) {
	t.Run("GIVEN 项目内 Go 文件是指向项目外的软链接", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")
		outsideDir := filepath.Join(tmpDir, "outside")
		outsideFile := filepath.Join(outsideDir, "outside.go")
		linkFile := filepath.Join(projectDir, "linked.go")

		Must(t, func() error {
			if err := os.MkdirAll(projectDir, 0o755); err != nil {
				return err
			}
			if err := os.MkdirAll(outsideDir, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte("module example.com/project\n\ngo 1.26.0\n"), 0o644); err != nil {
				return err
			}
			if err := os.WriteFile(outsideFile, []byte("package outside\n\nfunc Outside(){\n}\n"), 0o644); err != nil {
				return err
			}
			return os.Symlink(outsideFile, linkFile)
		})

		wd := MustValue(t, os.Getwd)
		Must(t, func() error {
			return os.Chdir(projectDir)
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

			Then(t, "THEN 应拒绝软链接写出项目根目录",
				ExpectDo(func() error {
					return p.Run(context.Background())
				}, ErrorMatch(regexp.MustCompile("软链接目标超出项目根目录"))),
			)

			data := MustValue(t, func() ([]byte, error) {
				return os.ReadFile(outsideFile)
			})

			Then(t, "THEN 项目外软链接目标不应被格式化",
				Expect(strings.Contains(string(data), "func Outside(){\n"), Be(cmp.True())),
			)
		})
	})
}
