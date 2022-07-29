package b

import "golang.org/x/tools/go/vcs"

// B is a type for testing
type B string

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

type Third vcs.RepoRoot
