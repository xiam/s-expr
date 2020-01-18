package parser

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/sexpr/node"
)

func TestParserBuildTree(t *testing.T) {
	testCases := []struct {
		In string
	}{
		{
			In: `[]`,
		},
		{
			In: `1`,
		},
		{
			In: `1 3 3.4 5.6789`,
		},
		{
			In: `[1 2 3]`,
		},
		{
			In: "[1\n\t 2\n\n3\n]",
		},
		{
			In: `[1.2 2.4 3.44 5.678]`,
		},
		{
			In: `[1.2 2.4 3.44 5.678 [1 1.2 1.3 [2] 1.4] 4 5 [6 [7] 8] [9 10] [] 11 12 13]`,
		},
		{
			In: `[1 [ 1 2 3 ] 3] 4 [5 6] 7 8`,
		},
		{
			In: `[1 [ 1 2 3 ] 3]`,
		},
		{
			In: `(1 2 3)`,
		},
		{
			In: `() (1) ()`,
		},
		{
			In: `(1 2 [] [3[4[5]]] 6 (7))`,
		},
		{
			In: `[(1[2])]`,
		},
		{
			In: `([(1[2])]3)`,
		},
		{
			In: "(a		b c def GHIJ 1 1.23)",
		},
		{
			In: "(a :b :cdef :GHI [jkl		:hijk])",
		},
		{
			In: ":true\n\n\n\n:false\n:nil\na\nCBD",
		},
		{
			In: ":true\n\n\n\n:false\n:nil\na\nCBD",
		},
		{
			In: `"ABC		DEF	[] GHI :jkl mno" :aBC def ghij "foo BAR" # extra`,
		},
		{
			In: "\"ABC	\\n	DEF	[] GHI :jkl mno\" # AABBCBCC\n:aBC #def ghij\n \"foo\" # BAR",
		},
		{
			In: `{}`,
		},
		{
			In: `{:foo 1}`,
		},
		{
			In: `[{:foo 1}]`,
		},
		{
			In: "set {:foo 1\n:bar 2} {:baz {{{[(1)]}}}}",
		},
		{
			In: `(fn [a b c] [(print a b c)])`,
		},
		{
			In: `{:a 1.11 :b "STRING VALUE"} ()`,
		},
		{
			In: `(fn word [] [(print %1 %2 %3 %*)]) (word "a" "b" "c")`,
		},
		{
			In: `(fn [] [(print %1 %2 %3 %*)])`,
		},
		{
			In: `(print "hello world" "beautiful world!") (echo :brave :new :world)`,
		},
		{
			In: `(+ 1 2 3 4)`,
		},
		{
			In: `{+ 1 2 3 4}`,
		},
		{
			In: `[+ 1 2 3 4]`,
		},
	}

	for i := range testCases {
		root, err := Parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)

		node.Print(root)
		s := compileNode(root)
		log.Printf("compiled: %v", string(s))
		log.Printf("tree: %#v", root)
	}
}
