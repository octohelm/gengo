package generators

import (
	"github.com/go-courier/gengo/pkg/gengo"
	"github.com/go-courier/gengo/pkg/namer"
	"go/types"
	"io"
)

func NewDefaulterGen() gengo.Generator {
	return &defaulterGen{
		imports: namer.NewDefaultImportTracker(),
	}
}

type defaulterGen struct {
	imports namer.ImportTracker
}

func (defaulterGen) Name() string {
	return "defaulter"
}

func (d defaulterGen) Imports(c *gengo.Context) map[string]string {
	return d.imports.Imports()
}

func (d defaulterGen) Init(c *gengo.Context, w io.Writer) error {
	return nil
}

func (d defaulterGen) GenerateType(c *gengo.Context, t types.Type, w io.Writer) error {
	sw := gengo.NewSnippetWriter(w, namer.NameSystems{
		"raw": namer.NewRawNamer(c.Package.Pkg().Path(), d.imports),
	})

	args := map[string]interface{}{
		"type": t,
	}

	if err := sw.Do(`
func(v *{{ .type | raw }}) SetDefault() {
}
`, args); err != nil {
		return err
	}

	return nil
}
