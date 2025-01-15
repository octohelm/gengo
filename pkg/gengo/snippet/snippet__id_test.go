package snippet

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_pkgExpose(t *testing.T) {
	tpe := reflect.TypeOf(Test_pkgExpose)
	fmt.Println(tpe)
}
