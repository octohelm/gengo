package gengo_test

import (
	//"github.com/go-courier/schema/pkg/gengo"
	//"k8s.io/gengo/args"
	//"testing"

	"context"
	"github.com/go-courier/gengo/examples/defaulter-gen/generators"

	//enumerationgenerators "github.com/go-courier/schema/pkg/enumeration/generators"
	//jsonschemagenerators "github.com/go-courier/schema/pkg/jsonschema/generators"
	"github.com/go-courier/gengo/pkg/gengo"
	"testing"
)

func TestPkgGenerator(t *testing.T) {
	c, _ := gengo.NewContext(&gengo.GeneratorArgs{
		Inputs: []string{
			"github.com/go-courier/gengo/testdata/a/b",
			"github.com/go-courier/gengo/testdata/a/c",
		},
		OutputFileBaseName: "zz_generated",
	})

	if err := c.Execute(context.Background(), generators.NewDefaulterGen); err != nil {
		t.Fatal(err)
	}
}
