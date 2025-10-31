package format

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/octohelm/gengo/pkg/format/internal"
)

type Project struct {
	Entrypoint []string `arg:""`

	List  bool `flag:",omitzero" alias:"l"`
	Write bool `flag:",omitzero" alias:"w"`

	cwd string
}

func (p *Project) Init(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	p.cwd = cwd
	return nil
}

func (p *Project) Run(ctx context.Context) error {
	for _, entry := range p.Entrypoint {
		entry = filepath.Clean(entry)
		if !filepath.IsAbs(entry) {
			entry = filepath.Join(p.cwd, entry)
		}
		if err := filepath.WalkDir(entry, func(path string, d fs.DirEntry, err error) error {
			explicit := path == entry
			switch {
			case err != nil:
				return err
			case d.IsDir():
				if !explicit && shouldIgnore(path) {
					return filepath.SkipDir
				}
				// simply recurse into directories
				return nil
			case explicit:
				// non-directories given as explicit arguments are always formatted
			case !isGoFilename(d.Name()):
				return nil // skip walked non-Go files
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			if fileWeight(path, info) == exclusive {
				return nil
			}

			return p.process(path, info)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) process(path string, info os.FileInfo) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	opt := Options{}

	module := internal.LoadModule(filepath.Dir(path))
	if module != nil {
		opt = OptionsFromModFile(module.File)
	}

	formated, err := Source(data, opt)
	if err != nil {
		return err
	}

	if !bytes.Equal(data, formated) {
		if p.List {
			relPath, _ := filepath.Rel(p.cwd, path)
			fmt.Println(relPath)
		}
		if p.Write {
			return os.WriteFile(path, formated, info.Mode())
		}
	}

	return nil
}

const exclusive = -1

func fileWeight(path string, info fs.FileInfo) int64 {
	if info == nil {
		return exclusive
	}
	if info.Mode().Type() == fs.ModeSymlink {
		var err error
		info, err = os.Stat(path)
		if err != nil {
			return exclusive
		}
	}
	if !info.Mode().IsRegular() {
		// For non-regular files, FileInfo.Size is system-dependent and thus not a
		// reliable indicator of weight.
		return exclusive
	}
	return info.Size()
}

func shouldIgnore(path string) bool {
	switch filepath.Base(path) {
	case "vendor", "testdata":
		return true
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return false // unclear how this could happen; don't ignore in any case
	}
	m := internal.LoadModule(path)
	if m == nil {
		// no module file to declare ignore paths
		return false
	}
	relPath, err := filepath.Rel(m.Dir, path)
	if err != nil {
		return false // unclear how this could happen; don't ignore in any case
	}
	relPath = normalizePath(relPath)
	for _, ignore := range m.File.Ignore {
		if matchIgnore(ignore.Path, relPath) {
			return true
		}
	}
	return false
}

func isGoFilename(name string) bool {
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func normalizePath(path string) string {
	path = filepath.ToSlash(path) // ensure Windows support
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func matchIgnore(ignore, relPath string) bool {
	ignore, rooted := strings.CutPrefix(ignore, "./")
	ignore = normalizePath(ignore)
	// Note that we only match the directory to be ignored itself,
	// and not any directories underneath it.
	// This way, using `gofumpt -w ignored` allows `ignored/subdir` to be formatted.
	if rooted {
		return relPath == ignore
	}
	return strings.HasSuffix(relPath, ignore)
}
