package testingutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/octohelm/gengo/pkg/gengo"
)

const defaultModulePath = "example.com/gengo-test"

var chdirMu sync.Mutex

// Module 表示一个隔离的临时 Go module。
type Module struct {
	t testing.TB

	Dir        string
	ModulePath string
}

// NewModule 创建临时 Go module，并写入给定文件。
func NewModule(t testing.TB, files map[string]string) (*Module, error) {
	t.Helper()

	m := &Module{
		t:          t,
		Dir:        t.TempDir(),
		ModulePath: defaultModulePath,
	}

	root, err := repoRoot()
	if err != nil {
		return nil, err
	}

	if err := m.write("go.mod", strings.Join([]string{
		"module " + m.ModulePath,
		"",
		"go 1.26.2",
		"",
		"require github.com/octohelm/gengo v0.0.0",
		"",
		"replace github.com/octohelm/gengo => " + root,
		"",
	}, "\n")); err != nil {
		return nil, err
	}

	for name, content := range files {
		if err := m.write(name, content); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Entrypoint 返回临时模块中 relPkg 对应的绝对目录入口。
func (m *Module) Entrypoint(relPkg string) string {
	m.t.Helper()

	return filepath.Join(m.Dir, filepath.FromSlash(relPkg))
}

// ImportPath 返回临时模块中 relPkg 对应的 import path。
func (m *Module) ImportPath(relPkg string) string {
	m.t.Helper()

	if relPkg == "" || relPkg == "." {
		return m.ModulePath
	}
	return m.ModulePath + "/" + strings.Trim(strings.TrimPrefix(relPkg, "./"), "/")
}

// Generate 执行生成器，并返回本次生成的文件内容。
func (m *Module) Generate(args gengo.GeneratorArgs, generators ...gengo.Generator) (map[string]string, error) {
	m.t.Helper()

	if err := m.GenerateError(args, generators...); err != nil {
		return nil, err
	}

	if args.OutputFileBaseName == "" {
		args.OutputFileBaseName = "zz_generated"
	}

	return m.ReadGenerated(args.OutputFileBaseName + ".*.go")
}

// GenerateError 执行生成器，并返回执行错误。
func (m *Module) GenerateError(args gengo.GeneratorArgs, generators ...gengo.Generator) error {
	m.t.Helper()

	chdirMu.Lock()
	defer chdirMu.Unlock()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(m.Dir); err != nil {
		return err
	}
	defer func() {
		_ = os.Chdir(wd)
	}()

	executor, err := gengo.NewExecutor(&args)
	if err != nil {
		return err
	}

	if err := executor.Execute(context.Background(), generators...); err != nil {
		return err
	}
	return nil
}

// ReadGenerated 读取临时模块中匹配 pattern 的生成文件。
func (m *Module) ReadGenerated(pattern string) (map[string]string, error) {
	m.t.Helper()

	files := map[string]string{}
	if err := filepath.WalkDir(m.Dir, func(filename string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		matched, err := filepath.Match(pattern, filepath.Base(filename))
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(m.Dir, filename)
		if err != nil {
			return err
		}

		files[filepath.ToSlash(rel)] = string(data)
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func (m *Module) write(name string, content string) error {
	m.t.Helper()

	filename := filepath.Join(m.Dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", filepath.Dir(filename), err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入文件失败 %s: %w", filename, err)
	}
	return nil
}

func repoRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("无法定位 testingutil 源文件")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../../..")), nil
}
