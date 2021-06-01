package namer

import (
	gengotypes "github.com/go-courier/gengo/pkg/types"
	"go/types"
)

type Namer interface {
	Name(types.Type) string
}

type NameSystems map[string]Namer

type Names map[types.Type]string

func NewRawNamer(pkgPath string, tracker ImportTracker) Namer {
	return &rawNamer{pkgPath: pkgPath, tracker: tracker}
}

type rawNamer struct {
	pkgPath string
	tracker ImportTracker
	Names
}

func (n *rawNamer) Name(t types.Type) string {
	if n.Names == nil {
		n.Names = Names{}
	}

	if name, ok := n.Names[t]; ok {
		return name
	}

	var typeName gengotypes.TypeName

	switch x := t.(type) {
	case *types.Named:
		typeName = x.Obj()
	case gengotypes.TypeName:
		typeName = x
	}

	if typeName != nil {
		pkgPath := typeName.Pkg().Path()

		if pkgPath == n.pkgPath {
			return typeName.Name()
		} else {
			n.tracker.AddType(t)
			return n.tracker.LocalNameOf(pkgPath) + "." + typeName.Name()
		}
	}

	return t.String()
}
