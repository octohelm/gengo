package testingutil_test

import (
	"fmt"

	"github.com/octohelm/x/cmp"

	"github.com/octohelm/gengo/pkg/gengo/testingutil"
)

func ExampleFile() {
	files := map[string]string{
		"sample/zz_generated.demo.go": "package sample\nfunc Demo() {}\n",
	}

	err := testingutil.File("sample/zz_generated.demo.go",
		testingutil.Contains("package sample"),
		testingutil.NotContains("func Missing()"),
		testingutil.Count("func ", cmp.Eq(1)),
	)(files)

	fmt.Println(err == nil)

	// Output:
	// true
}
