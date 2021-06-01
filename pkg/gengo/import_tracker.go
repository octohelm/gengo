package gengo

import (
	"github.com/go-courier/gengo/pkg/namer"
	"go/types"
)

func NewImportTracker(typesToAdd ...types.Type) namer.ImportTracker {
	tracker := namer.NewDefaultImportTracker()

	for i := range typesToAdd {
		tracker.AddType(typesToAdd[i])
	}

	return tracker

}
