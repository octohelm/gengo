package format

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/octohelm/gengo/pkg/format/internal"
)

// Project 会遍历一个或多个入口路径，并格式化遇到的 Go 文件。
type Project struct {
	Entrypoint []string `arg:""`

	List  bool `flag:",omitzero" alias:"l"`
	Write bool `flag:",omitzero" alias:"w"`

	cwd         string
	projectRoot string
}

// Init 记录当前工作目录和项目根目录，供后续限制执行边界和输出相对路径。
func (p *Project) Init(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	p.cwd = cwd

	if mod := internal.LoadModule(cwd); mod != nil {
		p.projectRoot = mod.Dir
	} else {
		p.projectRoot = cwd
	}

	return nil
}

// Run 通过 go/packages 加载每个入口对应的包，并格式化其中的 Go 文件。
//
// .go 文件会直接处理；其他入口（目录/包路径，支持 ... 递归）使用 packages.Load
// 进行解析，与 go build / go fmt 的行为一致。
func (p *Project) Run(ctx context.Context) error {
	for _, entry := range p.Entrypoint {
		// 单个 .go 文件：直接处理，不作为 package pattern 传入
		if strings.HasSuffix(entry, ".go") {
			path := p.resolveFileEntry(entry)
			if !withinRoot(p.projectRoot, path) {
				return fmt.Errorf("入口 %q 超出项目根目录 %q", path, p.projectRoot)
			}
			if err := p.processFile(path); err != nil {
				return err
			}
			continue
		}

		c := &packages.Config{
			Mode: packages.NeedFiles,
			Dir:  p.cwd,
		}
		pkgs, err := packages.Load(c, entry)
		if err != nil {
			return err
		}

		for _, pkg := range pkgs {
			if len(pkg.Errors) > 0 {
				for _, e := range pkg.Errors {
					fmt.Fprintln(os.Stderr, "[warning]", e.Pos, e.Msg)
				}
			}
			for _, filename := range pkg.GoFiles {
				if err := p.processFile(filename); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// resolveFileEntry 将 .go 文件入口转为绝对路径。
func (p *Project) resolveFileEntry(entry string) string {
	entry = filepath.Clean(entry)
	if !filepath.IsAbs(entry) {
		entry = filepath.Join(p.cwd, entry)
	}
	return entry
}

// processFile 对单个 Go 文件执行格式化，并按 List / Write 输出或回写。
func (p *Project) processFile(filename string) error {
	if !withinRoot(p.projectRoot, filename) {
		return fmt.Errorf("路径 %q 超出项目根目录 %q", filename, p.projectRoot)
	}

	// 检查软链接目标是否超出项目根目录
	fi, err := os.Lstat(filename)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		realPath, err := filepath.EvalSymlinks(filename)
		if err != nil {
			return err
		}
		if !withinRoot(p.projectRoot, realPath) {
			return fmt.Errorf("路径 %q 的软链接目标超出项目根目录 %q", filename, p.projectRoot)
		}
		fi, err = os.Stat(filename)
		if err != nil {
			return err
		}
	}

	if !fi.Mode().IsRegular() {
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	opt := Options{}
	module := internal.LoadModule(filepath.Dir(filename))
	if module != nil {
		opt = OptionsFromModFile(module.File)
	}

	formated, err := Source(data, opt)
	if err != nil {
		return err
	}

	if !bytes.Equal(data, formated) {
		if p.List {
			relPath, _ := filepath.Rel(p.cwd, filename)
			fmt.Println(relPath)
		}
		if p.Write {
			return os.WriteFile(filename, formated, fi.Mode())
		}
	}

	return nil
}

func withinRoot(root string, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || rel == "" || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}
