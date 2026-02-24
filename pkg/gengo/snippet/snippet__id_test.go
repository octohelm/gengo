package snippet

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_pkgExpose(t *testing.T) {
	tpe := reflect.TypeFor[func(t *testing.T)]()
	fmt.Println(tpe)
}
