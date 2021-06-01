package namer

import (
	"go/token"
	"go/types"
	"regexp"
	"strings"

	gengotypes "github.com/go-courier/gengo/pkg/types"
)

type ImportTracker interface {
	AddType(types.Type)

	LocalNameOf(packagePath string) string
	PathOf(localName string) (string, bool)

	Imports() map[string]string
}

type defaultImportTracker struct {
	pathToName map[string]string
	nameToPath map[string]string
}

func NewDefaultImportTracker() ImportTracker {
	return &defaultImportTracker{
		pathToName: map[string]string{},
		nameToPath: map[string]string{},
	}
}

var reInvalidIDChar = regexp.MustCompile(`^[A-Za-z0-9_]`)

func (tracker *defaultImportTracker) AddType(t types.Type) {
	if canPkg, ok := t.(gengotypes.TypeName); ok {
		path := canPkg.Pkg().Path()

		if _, ok := tracker.pathToName[path]; ok {
			return
		}

		localName := golangTrackerLocalName(path)

		tracker.nameToPath[localName] = path
		tracker.pathToName[path] = localName
	}
}

func golangTrackerLocalName(name string) string {
	name = strings.Replace(name, "/", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "-", "_", -1)

	if token.Lookup(name).IsKeyword() {
		name = "_" + name
	}

	return name
}

func (tracker *defaultImportTracker) Imports() map[string]string {
	return tracker.nameToPath
}

func (tracker *defaultImportTracker) LocalNameOf(path string) string {
	return tracker.pathToName[path]
}

func (tracker *defaultImportTracker) PathOf(localName string) (string, bool) {
	name, ok := tracker.nameToPath[localName]
	return name, ok
}
