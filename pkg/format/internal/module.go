package internal

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"

	"github.com/octohelm/x/sync/singleflight"
)

var cachedModuleByDir singleflight.GroupValue[string, *CachedModule]

type CachedModule struct {
	Dir  string
	File *modfile.File
}

func LoadModule(dir string) *CachedModule {
	mod, _, _ := cachedModuleByDir.Do(dir, func() (*CachedModule, error) {
		goModFilename := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(goModFilename)
		if errors.Is(err, fs.ErrNotExist) {
			parent := filepath.Dir(dir)
			if parent == "." {
				panic("loadModule was not given an absolute path?")
			}
			if parent == dir {
				// reached the filesystem root
				return nil, nil
			}
			// try the parent directory
			return LoadModule(parent), nil
		}
		if err != nil {
			return nil, nil
		}
		file, err := modfile.Parse(goModFilename, data, nil)
		if err != nil {
			return nil, nil
		}
		return &CachedModule{
			Dir:  dir,
			File: file,
		}, nil
	})
	return mod
}
