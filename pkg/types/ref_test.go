package types

import (
	"reflect"
	"testing"

	testingx "github.com/octohelm/x/testing"
)

type List[T any] struct {
	Items []T `json:"items"`
}

func TestRef(t *testing.T) {
	tpe := reflect.TypeOf(List[string]{})

	ref, _ := ParseRef(Ref(tpe.PkgPath(), tpe.Name()).String())
	testingx.Expect(t, ref.Pkg().Path(), testingx.Be("github.com/octohelm/gengo/pkg/types"))
	testingx.Expect(t, ref.Name(), testingx.Be("List[string]"))
}

func TestTypeRef(t *testing.T) {
	tpe := reflect.TypeOf(List[List[string]]{})

	x := Ref(tpe.PkgPath(), tpe.Name()).String()

	ref, err := ParseTypeRef(x)
	testingx.Expect(t, err, testingx.BeNil[error]())
	testingx.Expect(t, ref.String(), testingx.Be("github.com/octohelm/gengo/pkg/types.List[github.com/octohelm/gengo/pkg/types.List[string]]"))
}
