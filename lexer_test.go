package sexpr

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexerScanner(t *testing.T) {

	testCases := []string{
		`1`,

		`+ 1 1 1 1`,

		`[ [ [] ] [] []]`,

		`(+ 1 2 3)`,

		`(- 1 2 3)`,

		`(foo a b c-d-e-f "ghi")`,

		`(foo
			a :b
			c-d-e-f
			"g
			hi"
		)`,

		`(set foo (+ 3 3))`,

		`(get foo)`,

		`(fn sum [ a b ] [
			(+ a b)
		])`,

		`
		(fn sum

			[] [
				(print %1 %2 %*)
			]

			[a b c] [
				[ 4 4 4 ]
				[
					(print [ [1] [2] ])
					(* 2 a)
					(* 3 b)
					(* 4 c)
				]
			]

			[a b] [
				(+ a b)
				(- a b)
				(* a b)
			]

			[a] [
				(* a a)
				(- a a)
				(+ a a)
			]

			[...n] [
				(set x "xxxx" :xxxx)
				(* (get n 1 2 3) (get n 2))
				(* (get n 1) (get n 2))
			]
		)
		`,

		`(
		set
			a 1
			b 3
			c [ 4 4 4 ]
			d (fn [a] [
					[1]
				])
			e {
				:a 1
				:b 2
				:c 3
			}
		)`,

		`(
			"hello world!" "brave new " :world
		)`,

		`(+ 1 2 3 4)`,
	}

	{
		for i := range testCases {
			tokens, err := tokenize([]byte(testCases[i]))
			log.Printf("tokens: %v", tokens)

			assert.NotNil(t, tokens)
			assert.NoError(t, err)
		}
	}
}

func TestLexerTokenize(t *testing.T) {

	testCases := []struct {
		In  string
		Out []tokenType
	}{
		{
			`1`,
			[]tokenType{
				tokenInteger,
				tokenEOF,
			},
		},
		{
			`+
			1`,
			[]tokenType{
				tokenString,
				tokenNewLine,
				tokenWhitespace,
				tokenInteger,
				tokenEOF,
			},
		},
		{
			`(+
			[1
			{}])`,
			[]tokenType{
				tokenOpenExpression,
				tokenString,
				tokenNewLine,
				tokenWhitespace,
				tokenOpenList,
				tokenInteger,
				tokenNewLine,
				tokenWhitespace,
				tokenOpenMap,
				tokenCloseMap,
				tokenCloseList,
				tokenCloseExpression,
				tokenEOF,
			},
		},
	}

	getTokenTypes := func(tokens []token) []tokenType {
		tt := make([]tokenType, 0, len(tokens))
		for i := range tokens {
			tt = append(tt, tokens[i].tt)
		}
		return tt
	}

	{
		for i := range testCases {
			tokens, err := tokenize([]byte(testCases[i].In))

			assert.NotNil(t, tokens)
			assert.NoError(t, err)

			assert.Equal(t, testCases[i].Out, getTokenTypes(tokens))
		}
	}

}

func TestLexerColumnAndLines(t *testing.T) {

	testCases := []struct {
		In  string
		Pos [][2]int
	}{
		{
			"",
			[][2]int{
				{1, 1},
			},
		},
		{
			"1",
			[][2]int{
				{1, 1}, {1, 2},
			},
		},
		{
			"\n\n\n\n",
			[][2]int{
				{1, 1},
				{2, 1},
				{3, 1},
				{4, 1},
				{5, 1},
			},
		},
		{
			"\n\n\nABCDF efgh\n",
			[][2]int{
				{1, 1},
				{2, 1},
				{3, 1},
				{4, 1}, {4, 6}, {4, 7}, {4, 11},
				{5, 1},
			},
		},
		{
			"1\n\n\t\t23456",
			[][2]int{
				{1, 1}, {1, 2},
				{2, 1},
				{3, 1}, {3, 3}, {3, 8},
			},
		},
	}

	getTokenPositions := func(tokens []token) [][2]int {
		ret := make([][2]int, 0, len(tokens))
		for i := range tokens {
			ret = append(ret, [2]int{tokens[i].line, tokens[i].col})
		}
		return ret
	}

	{
		for i := range testCases {
			tokens, err := tokenize([]byte(testCases[i].In))

			assert.NotNil(t, tokens)
			assert.NoError(t, err)

			assert.Equal(t, testCases[i].Pos, getTokenPositions(tokens))
		}
	}

}
