package namer

import (
	"slices"
	"strconv"
	"strings"

	"github.com/octohelm/gengo/pkg/camelcase"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	"golang.org/x/tools/go/packages"
)

type ImportTracker interface {
	AddType(o gengotypes.TypeName)

	LocalNameOf(packagePath string) string
	PathOf(localName string) (string, bool)

	Imports() map[string]string
}

type defaultImportTracker struct {
	pathToName map[string]string
	nameToPath map[string]string
	checkStd   bool
}

func NewDefaultImportTracker() ImportTracker {
	return &defaultImportTracker{
		pathToName: map[string]string{},
		nameToPath: map[string]string{},
		checkStd:   true,
	}
}

var std = &defaultImportTracker{
	pathToName: map[string]string{},
	nameToPath: map[string]string{},
}

func init() {
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		panic(err)
	}
	for _, p := range pkgs {
		std.add(p.PkgPath)
	}
}

func (tracker *defaultImportTracker) AddType(o gengotypes.TypeName) {
	tracker.add(o.Pkg().Path())
}

func (tracker *defaultImportTracker) add(path string) {
	if _, ok := tracker.pathToName[path]; ok {
		return
	}

	parts := strings.Split(path, "/")

	for i := range len(parts) {
		localName := golangTrackerLocalName(parts, i+1)

		if tracker.checkStd {
			if p, ok := std.nameToPath[localName]; ok && p != path {
				continue
			}
		}

		if _, ok := tracker.nameToPath[localName]; !ok {
			tracker.nameToPath[localName] = path
			tracker.pathToName[path] = localName
			break
		}
	}
}

func toLocalName(parts ...string) string {
	return strings.ToLower(camelcase.LowerCamelCase(strings.Join(parts, "")))
}

func golangTrackerLocalName(pathSegments []string, n int) string {
	if len(pathSegments) == 1 {
		return toLocalName(pathSegments[0])
	}

	if n == 1 {
		// first good check
		if i := slices.Index(pathSegments, "domain"); i > 0 && i+1 < len(pathSegments) {
			return toLocalName(pathSegments[i+1:]...)
		}
		if i := slices.Index(pathSegments, "apis"); i > 0 && i+1 < len(pathSegments) {
			return toLocalName(pathSegments[i+1:]...)
		}
	}

	parts := make([]string, 0)

	count := 0
	for _, seg := range slices.Backward(pathSegments) {
		if count >= n {
			break
		}

		// vNumber
		if strings.HasPrefix(seg, "v") {
			if _, err := strconv.ParseInt(seg[1:], 10, 64); err == nil {
				parts = append(parts, seg)
				continue
			}
		}
		parts = append(parts, seg)

		count++
	}

	slices.Reverse(parts)

	return toLocalName(parts...)
}

func (tracker *defaultImportTracker) Imports() map[string]string {
	return tracker.pathToName
}

func (tracker *defaultImportTracker) LocalNameOf(path string) string {
	return tracker.pathToName[path]
}

func (tracker *defaultImportTracker) PathOf(localName string) (string, bool) {
	name, ok := tracker.nameToPath[localName]
	return name, ok
}
