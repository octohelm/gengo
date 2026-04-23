package gengo_test

import (
	"context"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	. "github.com/octohelm/x/testing/v2"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

type pkgGenerator struct{}

func (*pkgGenerator) Name() string {
	return "defaulter"
}

func (*pkgGenerator) GenerateType(c gengo.Context, named *types.Named) error {
	c.RenderT("type @NameGenerated struct{}\n", snippet.IDArg("NameGenerated", named.Obj().Name()+"Generated"))
	return nil
}

func TestPkgGenerator(t *testing.T) {
	outputBaseName := "zz_generated_api_test"
	t.Cleanup(func() {
		for _, dir := range []string{"testdata/runtime/b", "testdata/runtime/c"} {
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
		return gengo.NewExecutor(&gengo.GeneratorArgs{
			Entrypoint: []string{
				"github.com/octohelm/gengo/pkg/gengo/testdata/runtime/b",
				"github.com/octohelm/gengo/pkg/gengo/testdata/runtime/c",
			},
			OutputFileBaseName: outputBaseName,
			All:                true,
		})
	})

	Must(t, func() error {
		return c.Execute(context.Background(), &pkgGenerator{})
	})
}
