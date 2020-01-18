package ast

import (
	"fmt"
	"strings"
)

// Print displays a human-readable representation of a node
func Print(n *Node) {
	printLevel(n, 0)
}

func printLevel(n *Node, level int) {
	if n == nil {
		fmt.Printf("()\n")
		return
	}
	indent := strings.Repeat("    ", level)
	fmt.Printf("%s(%s): ", indent, n.Type())
	switch n.Type() {

	case NodeTypeExpression, NodeTypeList, NodeTypeMap:
		fmt.Printf("%v\n", n.Token())
		list := n.List()
		for i := range list {
			printLevel(list[i], level+1)
		}

	default:
		fmt.Printf("%#v (%v)\n", n.Value(), n.Token())
	}
}

// Encode transform a node into text representation
func Encode(n *Node) []byte {
	return encodeNodeLevel(n, 0)
}

func encodeNodeLevel(n *Node, level int) []byte {
	if n == nil {
		return []byte("()")
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
		if level == 0 {
			return []byte(strings.Join(nodes, " "))
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

	default:
		return []byte(n.Encode())
	}
}
