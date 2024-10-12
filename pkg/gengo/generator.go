package gengo

import (
	"go/types"

	"errors"
)

var (
	ErrSkip   = errors.New("skip")
	ErrIgnore = errors.New("ignore")
)

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
	// GenerateType do generate for each named type
	GenerateType(Context, *types.Named) error
}

type GeneratorNewer interface {
	// New generator
	New(c Context) Generator
}

type GeneratorCreator interface {
	Init(Context, Generator, ...GeneratorPostInit) (Generator, error)
}

type GeneratorPostInit = func(g Generator, sw SnippetWriter) error
