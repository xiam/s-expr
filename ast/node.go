package ast

import (
	"errors"
	"fmt"

	"github.com/xiam/sexpr/lexer"
)

// NodeType represents the type of the AST node
type NodeType uint8

// Node types
const (
	NodeTypeValue = iota
	NodeTypeList
	NodeTypeMap
	NodeTypeExpression
)

func (nt NodeType) String() string {
	s, ok := nodeTypeName[nt]
	if ok {
		return s
	}
	return ""
}

var nodeTypeName = map[NodeType]string{
	NodeTypeExpression: "expression",
	NodeTypeList:       "list",
	NodeTypeMap:        "map",
	NodeTypeValue:      "value",
}

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

// New creates and returns an orphaned node based on the given token
func New(tok *lexer.Token, v interface{}) *Node {
	return newNode(NodeTypeValue, tok, v)
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
func (n *Node) PushValue(tok *lexer.Token, v interface{}) (*Node, error) {
	node := New(tok, v)
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
	return n.v
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
	if _, ok := n.v.([]*Node); ok {
		n.v = append(n.v.([]*Node), node)
		node.p = n
		return nil
	}
	return errors.New("node can't accept children")
}
