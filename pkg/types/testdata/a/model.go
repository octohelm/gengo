package a

import "time"

// String 是函数结果推导使用的本地命名类型。
type String string

// FakeBool
// openapi:type boolean
type FakeBool int

func (FakeBool) OpenAPISchemaType() []string { return []string{"boolean"} }

type Gender int

const (
	GENDER_UNKNOWN Gender = iota
	GENDER__MALE          // 男
	GENDER__FEMALE        // 女
)

type TimeAlias = time.Time

// Struct
type Struct struct {
	// StructID
	ID    **string
	Slice []float64
}

type Node struct {
	Children []*Node
}

type ListNode struct{}

func (req *ListNode) ResponseData() *Node {
	return new(Node)
}
