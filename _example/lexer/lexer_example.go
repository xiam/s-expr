package main

import (
	"fmt"
	"log"

	"github.com/xiam/sexpr/lexer"
)

func main() {
	input := `
		(fn_a # comment
			(fn_b [89 :A :B [67 3.27]])
			(fn_c 66 3 53 "Hello world!")
		)
	`

	tokens, err := lexer.Tokenize([]byte(input))
	if err != nil {
		log.Fatal("lexer.Tokenize:", err)
	}

	for i, tok := range tokens {
		line, col := tok.Pos()
		lexeme := tok.Text()
		tt := tok.Type().String()

		fmt.Printf("token[%d] (type: %v, line: %d, col: %d)\n\t-> %q\n\n", i, tt, line, col, lexeme)
	}
}
