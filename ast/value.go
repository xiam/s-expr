package ast

import (
	"fmt"
)

// Valuer represents a value interface
type Valuer interface {
	Type() NodeType
	Value() interface{}
	Encode() string
}

type nodeValue struct {
	t NodeType
	v interface{}
}

func newNodeValue(t NodeType, v interface{}) *nodeValue {
	return &nodeValue{
		t: t,
		v: v,
	}
}

func (n *nodeValue) Type() NodeType {
	return n.t
}

func (n *nodeValue) Value() interface{} {
	return n.v
}

func (n *nodeValue) Encode() string {
	switch n.t {
	case NodeTypeInt:
		return fmt.Sprintf("%d", n.v)
	case NodeTypeFloat:
		return fmt.Sprintf("%v", n.v)
	case NodeTypeSymbol:
		return fmt.Sprintf("%s", n.v)
	case NodeTypeAtom:
		return fmt.Sprintf("%s", n.v)
	case NodeTypeString:
		return fmt.Sprintf("%q", n.v)
	}

	panic("unreachable")
}

// NewStringValue creates a node of type string and sets it to the given value
func NewStringValue(v string) Valuer {
	return newNodeValue(NodeTypeString, v)
}

// NewFloatValue creates a node of type float and sets it to the given value
func NewFloatValue(v float64) Valuer {
	return newNodeValue(NodeTypeFloat, v)
}

// NewIntValue creates a node of type node and sets it to the given value
func NewIntValue(v int64) Valuer {
	return newNodeValue(NodeTypeInt, v)
}

// NewAtomValue creates a node of type atom and sets it to the given value
func NewAtomValue(v string) Valuer {
	return newNodeValue(NodeTypeAtom, v)
}

// NewSymbolValue creates a node of type symbol and sets it to the given value
func NewSymbolValue(v string) Valuer {
	return newNodeValue(NodeTypeSymbol, v)
}

var _ = Valuer(&nodeValue{})
