package gengo

import (
	"github.com/pkg/errors"
	"go/types"
)

var Skip = errors.New("generate skip")

type GeneratorArgs struct {
	// Entrypoint should be import path or valid related dir path
	Entrypoint []string
	// OutputFileBaseName is the prefix of generated filename
	OutputFileBaseName string
	// Globals contains tags for all pkgs
	Globals map[string][]string
}

type Generator interface {
	// Name generator name
	Name() string
	// New generator
	New(c Context) Generator
	// GenerateType do generate for each named type
	GenerateType(Context, *types.Named) error
}

type GeneratorTypeFilter interface {
	FilterType(Context, *types.Named) bool
}

type GeneratorCreator interface {
	Init(Context, Generator, ...GeneratorPostInit) (Generator, error)
}

type GeneratorPostInit = func(g Generator, sw SnippetWriter) error
