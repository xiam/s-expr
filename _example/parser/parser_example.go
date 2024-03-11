package main

import (
	"log"

	"github.com/xiam/s-expr/ast"
	"github.com/xiam/s-expr/parser"
)

func main() {
	input := `(fn_a (fn_b [89 :A :B [67 3.27]]) (fn_c 66 3 53 "Hello world!" 😊))`

	root, err := parser.Parse([]byte(input))
	if err != nil {
		log.Fatal("parser.Parse:", err)
	}

	ast.Print(root)
}
