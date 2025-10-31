package gengo_test

import (
	"reflect"
	"testing"

	testingx "github.com/octohelm/x/testing"

	"github.com/octohelm/gengo/pkg/gengo"
	gengotypes "github.com/octohelm/gengo/pkg/types"
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
