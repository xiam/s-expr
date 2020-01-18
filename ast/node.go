package ast

import (
	"fmt"

	"github.com/xiam/sexpr/lexer"
)

type NodeType uint8

const (
	NodeTypeExpression NodeType = iota
	NodeTypeList
	NodeTypeValue
	NodeTypeMap
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

func New(tok *lexer.Token, v interface{}) *Node {
	return newNode(NodeTypeValue, tok, v)
}

func NewExpression(tok *lexer.Token) *Node {
	return newNode(NodeTypeExpression, tok, []*Node{})
}

func NewMap(tok *lexer.Token) *Node {
	return newNode(NodeTypeMap, tok, []*Node{})
}

func NewList(tok *lexer.Token) *Node {
	return newNode(NodeTypeList, tok, []*Node{})
}

func (n *Node) push(node *Node) {
	n.v = append(n.v.([]*Node), node)
	node.p = n
}

func (n *Node) AppendChild(node *Node) {
	n.push(node)
}

func (n *Node) Push(tok *lexer.Token, v interface{}) *Node {
	node := New(tok, v)
	n.push(node)
	return node
}

func (n *Node) PushExpression(tok *lexer.Token) *Node {
	node := NewExpression(tok)
	n.push(node)
	return node
}

func (n *Node) PushList(tok *lexer.Token) *Node {
	node := NewList(tok)
	n.push(node)
	return node
}

func (n *Node) PushMap(tok *lexer.Token) *Node {
	node := NewMap(tok)
	n.push(node)
	return node
}

func (n Node) Token() *lexer.Token {
	return n.tok
}

func (n Node) Type() NodeType {
	return n.nt
}

func (n Node) Value() interface{} {
	return n.v
}

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
