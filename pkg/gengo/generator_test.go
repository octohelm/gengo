package gengo_test

import (
	"context"

	"testing"

	"github.com/go-courier/gengo/pkg/gengo"

	_ "github.com/go-courier/gengo/examples/defaulter-gen/generators"
)

func TestPkgGenerator(t *testing.T) {
	c, _ := gengo.NewContext(&gengo.GeneratorArgs{
		Inputs: []string{
			"github.com/go-courier/gengo/testdata/a/b",
			"github.com/go-courier/gengo/testdata/a/c",
		},
		OutputFileBaseName: "zz_generated",
	})

	if err := c.Execute(context.Background(), gengo.GetRegisteredGenerators()...); err != nil {
		t.Fatal(err)
	}
}
