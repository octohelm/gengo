package gengo

import (
	"errors"
	"go/types"
)

var (
	ErrSkip   = errors.New("skip")
	ErrIgnore = errors.New("ignore")
)

type GeneratorArgs struct {
	// Globals contains tags for all pkgs
	Globals map[string][]string
	// Entrypoint should be import path or valid related dir path
	Entrypoint []string
	// OutputFileBaseName is the prefix of generated filename
	OutputFileBaseName string
	// All enabled, will process all deps
	All bool
}

type Generator interface {
	// Name generator name
	Name() string
	// GenerateType do generate for each named type
	GenerateType(Context, *types.Named) error
}

type AliasGenerator interface {
	// Name generator name
	Name() string

	GenerateAliasType(Context, *types.Alias) error
}

type GeneratorNewer interface {
	// New generator
	New(c Context) Generator
}

type GeneratorCreator interface {
	Init(Context, Generator, ...GeneratorPostInit) (Generator, error)
}

type GeneratorPostInit = func(g Generator, sw SnippetWriter) error
