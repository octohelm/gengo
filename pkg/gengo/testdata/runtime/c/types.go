package c

type Object interface {
	DeepCopyObject() Object
}

// Runtime object.
// +gengo:deepcopy
// +gengo:deepcopy:interfaces=Object
type KubePkg struct {
	Spec KubePkgSpec
}

// +gengo:deepcopy
type KubePkgSpec struct {
	Version string
}

type TimeAlias = string
