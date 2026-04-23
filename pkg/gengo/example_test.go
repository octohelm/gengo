package gengo_test

import (
	"context"
	"fmt"
	"go/types"
	"os"
	"path/filepath"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/gengo/pkg/gengo/snippet"
)

type exampleGenerator struct {
	name string
}

func (g *exampleGenerator) Name() string {
	return g.name
}

func (g *exampleGenerator) GenerateType(c gengo.Context, named *types.Named) error {
	c.RenderT("type @NameGenerated struct{}\n", snippet.IDArg("NameGenerated", named.Obj().Name()+"Generated"))
	return nil
}

func ExampleRegister() {
	_ = gengo.Register(&exampleGenerator{name: "example"})
	err := gengo.Register(&exampleGenerator{name: "example"})
	fmt.Println(err != nil)

	// Output:
	// true
}

func ExampleNewExecutor() {
	const outputBaseName = "zz_generated_example"

	executor, err := gengo.NewExecutor(&gengo.GeneratorArgs{
		Entrypoint:         []string{"github.com/octohelm/gengo/pkg/gengo/testdata/runtime/b"},
		OutputFileBaseName: outputBaseName,
		Force:              true,
	})
	if err != nil {
		panic(err)
	}

	if err := executor.Execute(context.Background(), &exampleGenerator{name: "defaulter"}); err != nil {
		panic(err)
	}
	defer func() {
		matches, _ := filepath.Glob(filepath.Join("testdata/runtime/b", outputBaseName+".*.go"))
		for _, filename := range matches {
			_ = os.Remove(filename)
		}
	}()
}

func ExampleIsGeneratorEnabled() {
	g := &exampleGenerator{name: "client"}

	fmt.Println(gengo.IsGeneratorEnabled(g, map[string][]string{
		"gengo:client:openapi": {"https://example.com/openapi.json"},
	}))
	fmt.Println(gengo.IsGeneratorEnabled(g, map[string][]string{
		"gengo:client": {"false"},
	}))

	// Output:
	// true
	// false
}

func ExamplePkgImportPathAndExpose() {
	pkgPath, expose := gengo.PkgImportPathAndExpose("github.com/acme/project/pkg.User")
	fmt.Println(pkgPath)
	fmt.Println(expose)

	// Output:
	// github.com/acme/project/pkg
	// User
}
