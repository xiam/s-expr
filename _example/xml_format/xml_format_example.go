package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/xiam/sexpr/ast"
	"github.com/xiam/sexpr/parser"
)

func printTree(node *ast.Node) {
	printIndentedTree(node, 0)
}

func printIndentedTree(node *ast.Node, indentationLevel int) {
	indent := strings.Repeat("  ", indentationLevel)
	if node.IsVector() {
		fmt.Printf("%s<%s>\n", indent, node.Type())
		children := node.List()
		for i := range children {
			printIndentedTree(children[i], indentationLevel+1)
		}
		fmt.Printf("%s</%s>\n", indent, node.Type())
		return
	}
	fmt.Printf("%s<%s>%v</%s>\n", indent, node.Type(), node.Value(), node.Type())
}

func main() {
	input := `(fn_a (fn_b [89 :A :B [67 3.27]]) (fn_c 66 3 53 "Hello world!" 😊))`

	root, err := parser.Parse([]byte(input))
	if err != nil {
		log.Fatal("parser.Parse:", err)
	}

	printTree(root)
}
