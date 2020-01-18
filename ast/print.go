package ast

import (
	"fmt"
	"strings"

	"github.com/xiam/sexpr/lexer"
)

// Print displays a human-readable representation of a node
func Print(n *Node) {
	printLevel(n, 0)
}

func printLevel(n *Node, level int) {
	if n == nil {
		fmt.Printf(":nil\n")
		return
	}
	indent := strings.Repeat("    ", level)
	fmt.Printf("%s(%s): ", indent, n.Type())
	switch n.Type() {

	case NodeTypeExpression, NodeTypeList, NodeTypeMap:
		fmt.Printf("(%v)\n", n.Token())
		list := n.List()
		for i := range list {
			printLevel(list[i], level+1)
		}

	case NodeTypeValue:
		fmt.Printf("%#v (%v)\n", n.Value(), n.Token())

	default:
		panic("unknown node type")
	}
}

// Encode transform a node into text representation
func Encode(n *Node) []byte {
	return encodeNodeLevel(n, 0)
}

func encodeNodeLevel(n *Node, level int) []byte {
	if n == nil {
		return []byte(":nil")
	}
	switch n.Type() {
	case NodeTypeMap:
		nodes := []string{}
		for i := range n.List() {
			nodes = append(nodes, string(encodeNodeLevel(n.List()[i], level+1)))
		}
		return []byte(fmt.Sprintf("{%s}", strings.Join(nodes, " ")))

	case NodeTypeList:
		nodes := []string{}
		for i := range n.List() {
			nodes = append(nodes, string(encodeNodeLevel(n.List()[i], level+1)))
		}
		return []byte(fmt.Sprintf("[%s]", strings.Join(nodes, " ")))

	case NodeTypeExpression:
		nodes := []string{}
		for i := range n.List() {
			nodes = append(nodes, string(encodeNodeLevel(n.List()[i], level+1)))
		}
		if level == 0 {
			return []byte(fmt.Sprintf("%s", strings.Join(nodes, " ")))
		}
		return []byte(fmt.Sprintf("(%s)", strings.Join(nodes, " ")))

	case NodeTypeValue:
		if n.Token().Is(lexer.TokenString) {
			return []byte(fmt.Sprintf("%q", n.Value()))
		}
		return []byte(fmt.Sprintf("%v", n.Value()))

	default:
		panic("unknown node type")
	}
}
