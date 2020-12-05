package ast

import (
	"errors"
	"fmt"

	"github.com/xiam/sexpr/lexer"
)

// Node represents leaf of the AST
type Node struct {
	p *Node

	nt  NodeType
	tok *lexer.Token
	v   interface{}
}

func newNode(nt NodeType, tok *lexer.Token, v interface{}) *Node {
	return &Node{
		nt:  nt,
		v:   v,
		tok: tok,
	}
}

// NewNode creates and returns an orphaned node based on the given token
func NewNode(tok *lexer.Token, v Valuer) *Node {
	return newNode(v.Type(), tok, v)
}

// NewExpression creates and returns a node of type "expression"
func NewExpression(tok *lexer.Token) *Node {
	return newNode(NodeTypeExpression, tok, []*Node{})
}

// NewMap creates and returns a node of type "map"
func NewMap(tok *lexer.Token) *Node {
	return newNode(NodeTypeMap, tok, []*Node{})
}

// NewList creates and returns a node of type "list"
func NewList(tok *lexer.Token) *Node {
	return newNode(NodeTypeList, tok, []*Node{})
}

// PushValue appends a new value to the node
func (n *Node) PushValue(tok *lexer.Token, v Valuer) (*Node, error) {
	node := NewNode(tok, v)
	if err := n.Push(node); err != nil {
		return nil, err
	}
	return node, nil
}

// PushExpression appends a new expression to the node
func (n *Node) PushExpression(tok *lexer.Token) (*Node, error) {
	node := NewExpression(tok)
	if err := n.Push(node); err != nil {
		return nil, err
	}
	return node, nil
}

// PushList appends a new list to the node
func (n *Node) PushList(tok *lexer.Token) (*Node, error) {
	node := NewList(tok)
	if err := n.Push(node); err != nil {
		return nil, err
	}
	return node, nil
}

// PushMap appends a new map to the node
func (n *Node) PushMap(tok *lexer.Token) (*Node, error) {
	node := NewMap(tok)
	if err := n.Push(node); err != nil {
		return nil, err
	}
	return node, nil
}

// Token returns the token associated to the node
func (n Node) Token() *lexer.Token {
	return n.tok
}

// Type returns the type of the node
func (n Node) Type() NodeType {
	return n.nt
}

// Value returns the value of the node
func (n Node) Value() interface{} {
	if n.v == nil {
		return nil
	}
	if _, ok := n.v.(Valuer); ok {
		return n.v.(Valuer).Value()
	}
	return n.v
}

// Encode returns the encoded value of the node
func (n Node) Encode() string {
	if n.v == nil {
		return ""
	}
	if _, ok := n.v.(Valuer); ok {
		return n.v.(Valuer).Encode()
	}
	return ""
}

// List returns all the children elements of the node
func (n *Node) List() []*Node {
	return n.v.([]*Node)
}

func (n Node) String() string {
	switch n.nt {
	case NodeTypeExpression, NodeTypeList, NodeTypeMap:
		return fmt.Sprintf("(%v)[%d]", nodeTypeName[n.nt], len(n.List()))
	}
	return fmt.Sprintf("(%v): %v", nodeTypeName[n.nt], n.Value())
}

// Push appends a child node to a parent node of type "expression", "map" or "list".
func (n *Node) Push(node *Node) error {
	if n.IsVector() {
		n.v = append(n.v.([]*Node), node)
		node.p = n
		return nil
	}
	return errors.New("nodes of type value can't accept children")
}

// IsValue returns true if the node is of type value
func (n *Node) IsValue() bool {
	return n.nt&nodeTypeValue > 0
}

// IsVector returns true if the node is of type vector
func (n *Node) IsVector() bool {
	return n.nt&nodeTypeVector > 0
}

func (n *Node) Parent() *Node {
	return n.p
}
