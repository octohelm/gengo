package gengo_test

import (
	"context"
	"testing"

	"github.com/octohelm/gengo/pkg/gengo"

	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/defaultergen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
)

func TestPkgGenerator(t *testing.T) {
	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint: []string{
			"../../testdata/a/b",
			//"github.com/octohelm/gengo/testdata/a/c",
		},
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Execute(context.Background(), gengo.GetRegisteredGenerators()...); err != nil {
		t.Fatal(err)
	}
}
