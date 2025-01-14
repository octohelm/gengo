package b

import (
	"golang.org/x/mod/module"
)

// B is a type for testing
type B string

// Obj some object
// [[doc/b.md]]
type Obj struct {
	// name
	// 姓名
	Name string
	SubObj
}

type SubObj struct {
	// Age
	Age int
}

type Third module.Version
