package main

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/go/packages"
	"os"
	"slices"
	"strings"
)

func main() {
	pkgs, err := packages.Load(nil, "std")
	if err != nil {
		panic(err)
	}

	b := bytes.NewBuffer(nil)

	for _, p := range pkgs {
		if slices.Contains(strings.Split(p.PkgPath, "/"), "vendor") {
			continue
		}

		if slices.Contains(strings.Split(p.PkgPath, "/"), "internal") {
			continue
		}

		_, _ = fmt.Fprintln(b, p.PkgPath)
	}

	_ = os.WriteFile("std.list", b.Bytes(), os.ModePerm)
}
