package gengo_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo"
)

import (
	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/defaultergen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
)

func TestPkgGenerator(t *testing.T) {
	outputBaseName := "zz_generated_api_test"
	t.Cleanup(func() {
		for _, dir := range []string{"../../testdata/a/b", "../../testdata/a/c"} {
			matches, err := filepath.Glob(filepath.Join(dir, outputBaseName+".*.go"))
			if err != nil {
				continue
			}
			for _, filename := range matches {
				_ = os.Remove(filename)
			}
		}
	})

	c := MustValue(t, func() (gengo.Executor, error) {
		return gengo.NewContext(&gengo.GeneratorArgs{
			Entrypoint: []string{
				"../../testdata/a/b",
				"github.com/octohelm/gengo/testdata/a/c",
			},
			OutputFileBaseName: outputBaseName,
			All:                true,
		})
	})

	Must(t, func() error {
		return c.Execute(context.Background(), gengo.GetRegisteredGenerators()...)
	})
}
