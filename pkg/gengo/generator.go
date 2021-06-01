package gengo

import (
	"go/types"
	"io"
)

type Generator interface {
	Name() string
	Init(*Context, io.Writer) error
	Imports(*Context) map[string]string
	GenerateType(*Context, types.Type, io.Writer) error
}

type GeneratorArgs struct {
	Inputs             []string
	OutputFileBaseName string
}
