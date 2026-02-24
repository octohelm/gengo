package sumfile_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/sumfile"
)

func TestFile(t *testing.T) {
	t.Run("GIVEN 一个 sumfile", func(t *testing.T) {
		t.Run("WHEN 序列化 Bytes", func(t *testing.T) {
			f := &sumfile.File{
				Data: map[string]string{
					"z.io/c": "hash-c",
					"a.io/b": "hash-b",
					"b.io/a": "hash-a",
				},
			}

			Then(t, "THEN 应按包路径稳定排序输出",
				Expect(string(f.Bytes()), Equal(""+
					"a.io/b hash-b\n"+
					"b.io/a hash-a\n"+
					"z.io/c hash-c\n")),
			)
		})

		t.Run("WHEN 查询摘要", func(t *testing.T) {
			t.Run("THEN nil Data 应返回空字符串", func(t *testing.T) {
				f := &sumfile.File{}

				Then(t, "缺失记录应为空",
					Expect(f.Sum("a.io/b"), Equal("")),
				)
			})

			t.Run("THEN 已存在记录应返回对应摘要", func(t *testing.T) {
				f := &sumfile.File{
					Data: map[string]string{
						"a.io/b": "hash-b",
					},
				}

				Then(t, "已记录摘要应可正确返回",
					Expect(f.Sum("a.io/b"), Equal("hash-b")),
				)
			})
		})
	})

	t.Run("GIVEN 一个临时模块目录", func(t *testing.T) {
		tmpDir := t.TempDir()

		t.Run("WHEN Save 后再 Load", func(t *testing.T) {
			original := &sumfile.File{
				Dir: tmpDir,
				Data: map[string]string{
					"github.com/example/z": "hash-z",
					"github.com/example/a": "hash-a",
				},
			}

			Must(t, func() error {
				return original.Save()
			})

			data := MustValue(t, func() ([]byte, error) {
				return os.ReadFile(filepath.Join(tmpDir, "gengo.sum"))
			})

			Then(t, "THEN 文件内容应按稳定顺序写出",
				Expect(string(data), Equal(""+
					"github.com/example/a hash-a\n"+
					"github.com/example/z hash-z\n")),
			)

			loaded := MustValue(t, func() (*sumfile.File, error) {
				return sumfile.Load(tmpDir)
			})

			Then(t, "THEN Load 后目录应正确",
				Expect(loaded.Dir, Equal(tmpDir)),
			)

			Then(t, "THEN Load 后数据应完整恢复",
				Expect(loaded.Data, Equal(original.Data)),
			)
		})

		t.Run("WHEN 读取包含无效行的 gengo.sum", func(t *testing.T) {
			Must(t, func() error {
				return os.WriteFile(filepath.Join(tmpDir, "gengo.sum"), []byte(""+
					"github.com/example/a hash-a\n"+
					"invalid-line\n"+
					"github.com/example/b hash-b extra-columns\n"), 0o644)
			})

			loaded := MustValue(t, func() (*sumfile.File, error) {
				return sumfile.Load(tmpDir)
			})

			Then(t, "THEN 仅包含至少两列的行会被读取",
				Expect(loaded.Data, Equal(map[string]string{
					"github.com/example/a": "hash-a",
					"github.com/example/b": "hash-b",
				})),
			)
		})

		t.Run("WHEN 读取不存在的 gengo.sum", func(t *testing.T) {
			missingDir := filepath.Join(tmpDir, "missing")

			Then(t, "THEN 应返回 not exist 错误",
				ExpectMust(func() error {
					_, err := sumfile.Load(missingDir)
					if err == nil {
						return errors.New("expected not exist error")
					}
					if !os.IsNotExist(err) {
						return err
					}
					return nil
				}),
			)
		})
	})
}
