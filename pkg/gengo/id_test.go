package gengo_test

import (
	"github.com/octohelm/gengo/pkg/gengo"
	gengotypes "github.com/octohelm/gengo/pkg/types"
	testingx "github.com/octohelm/x/testing"
	"reflect"
	"testing"
)

type List[T any] struct {
	Items []T `json:"items"`
}

func TestPkgImportPathAndExpose(t *testing.T) {
	tpe := reflect.TypeOf(List[string]{})

	pkg, expose := gengo.PkgImportPathAndExpose(gengotypes.Ref(tpe.PkgPath(), tpe.Name()).String())
	testingx.Expect(t, pkg, testingx.Be("github.com/octohelm/gengo/pkg/gengo_test"))
	testingx.Expect(t, expose, testingx.Be("List"))
}
