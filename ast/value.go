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
		return fmt.Sprintf("%f", n.v)
	case NodeTypeSymbol:
		return fmt.Sprintf("%s", n.v)
	case NodeTypeAtom:
		return fmt.Sprintf("%s", n.v)
	case NodeTypeString:
		return fmt.Sprintf("%q", n.v)
	}

	panic("unreachable")
}

// NewStringNode creates a node of type string and sets it to the given value
func NewStringNode(v string) Valuer {
	return newNodeValue(NodeTypeString, v)
}

// NewFloatNode creates a node of type float and sets it to the given value
func NewFloatNode(v float64) Valuer {
	return newNodeValue(NodeTypeFloat, v)
}

// NewIntNode creates a node of type node and sets it to the given value
func NewIntNode(v int64) Valuer {
	return newNodeValue(NodeTypeInt, v)
}

// NewAtomNode creates a node of type atom and sets it to the given value
func NewAtomNode(v string) Valuer {
	return newNodeValue(NodeTypeAtom, v)
}

// NewSymbolNode creates a node of type symbol and sets it to the given value
func NewSymbolNode(v string) Valuer {
	return newNodeValue(NodeTypeSymbol, v)
}

var _ = Valuer(&nodeValue{})
