package sexpr

import (
	"fmt"
	"log"
	//"strings"
	"testing"
	//"time"

	"github.com/stretchr/testify/assert"
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
			In: "(a    b c def GHIJ 1 1.23)",
		},
		{
			In: "(a :b :cdef :GHI [jkl    :hijk])",
		},
		{
			In: ":true\n\n\n\n:false\n:nil\na\nCBD",
		},
		{
			In: ":true\n\n\n\n:false\n:nil\na\nCBD",
		},
		{
			In: `"ABC    DEF  [] GHI :jkl mno" :aBC def ghij "foo BAR" # extra`,
		},
		{
			In: "\"ABC  \\n  DEF  [] GHI :jkl mno\" # AABBCBCC\n:aBC #def ghij\n \"foo\" # BAR",
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
		root, err := parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)

		printNode(root)
		s := compileNode(root)
		log.Printf("compiled: %v", string(s))
		log.Printf("tree: %#v", root)
	}
}

func TestParserEvaluate(t *testing.T) {
	testCases := []struct {
		In  string
		Out string
	}{
		{
			In:  `1`,
			Out: `[1]`,
		},
		{
			In:  `[]`,
			Out: `[[]]`,
		},
		{
			In:  `[1]`,
			Out: `[[1]]`,
		},
		{
			In:  `[ 3 2  1 ]`,
			Out: `[[3 2 1]]`,
		},
		{
			In:  `[  1       2 [ 4 5 [6 7 8]] 3]`,
			Out: `[[1 2 [4 5 [6 7 8]] 3]]`,
		},
		{
			In:  `{}`,
			Out: `[{}]`,
		},
		{
			In:  `{:a}`,
			Out: `[{:a :nil}]`,
		},
		{
			In:  `{ :a 1     }`,
			Out: `[{:a 1}]`,
		},
		{
			In:  `{:a 1 :b 2 :c 3 :e [1 2 3]}`,
			Out: `[{:a 1 :b 2 :c 3 :e [1 2 3]}]`,
		},
		{
			In:  `[{:a 1 :b 2 :c 3 :e [1 2 3]} [1 2 3] 4 :foo]`,
			Out: `[[{:a 1 :b 2 :c 3 :e [1 2 3]} [1 2 3] 4 :foo]]`,
		},
		{
			In:  `(1)`,
			Out: `[1]`,
		},
		{
			In:  `( [1  2  3 ] )`,
			Out: `[[1 2 3]]`,
		},
		{
			In:  `(:nil)`,
			Out: `[:nil]`,
		},
		{
			In:  `(:hello)`,
			Out: `[:hello]`,
		},
		{
			In:  `(([1 2 3 {:a 4}]))`,
			Out: `[[1 2 3 {:a 4}]]`,
		},
		{
			In:  `(nop)`,
			Out: `[:nil]`,
		},
		{
			In:  `(print "hello world!" " beautiful world!")`,
			Out: `[:nil]`,
		},
		{
			In:  `(echo "hello world!" "beautiful world!"  1    2 )`,
			Out: `[["hello world!" "beautiful world!" 1 2]]`,
		},
		{
			In:  `(+ 1 2 3 4)`,
			Out: `[10]`,
		},
		{
			In:  `(+ (+ 1 2 3 4))`,
			Out: `[10]`,
		},
		{
			In:  `(+ (+ 1 2 3 4) 10)`,
			Out: `[20]`,
		},
		{
			In:  `((echo "+") (+ 1 2 3 4) 10 (15) 16)`,
			Out: `[51]`,
		},
		{
			In:  `(= 2 3)`,
			Out: `[:false]`,
		},
		{
			In:  `(= 1 1)`,
			Out: `[:true]`,
		},
		{
			In:  `(= 1 1 1 1 1 1 1)`,
			Out: `[:true]`,
		},
		{
			In:  `(= 1 1 1 1 1 2 14)`,
			Out: `[:false]`,
		},
	}

	RegisterPrefix("+", func(ctx *Context) (*Value, error) {
		defer ctx.Exit(nil)

		result := int64(0)
		for ctx.Next() {
			value := ctx.Argument()
			result += value.Int()
		}

		value, err := NewValue(result)
		if err != nil {
			return nil, err
		}

		ctx.Yield(value)
		return nil, nil
	})

	RegisterPrefix("echo", func(ctx *Context) (*Value, error) {
		defer ctx.Exit(nil)

		for ctx.Next() {
			value := ctx.Argument()
			ctx.Yield(value)
		}
		return nil, nil
	})

	RegisterPrefix("=", func(ctx *Context) (*Value, error) {
		defer ctx.Exit(nil)

		var first *Value
		for ctx.Next() {
			value := ctx.Argument()

			if first == nil {
				first = value
				continue
			}

			if (*first).raw() != value.raw() {
				ctx.Yield(False)
				return nil, nil
			}
		}

		ctx.Yield(True)
		return nil, nil
	})

	RegisterPrefix("nop", func(ctx *Context) (*Value, error) {
		defer ctx.Exit(nil)

		return nil, nil
	})

	RegisterPrefix("print", func(ctx *Context) (*Value, error) {
		defer ctx.Exit(nil)

		for ctx.Next() {
			value := ctx.Argument()
			fmt.Printf("%s", value.raw())
		}
		return nil, nil
	})

	for i := range testCases {
		root, err := parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)

		printNode(root)

		_, result, err := eval(root)
		assert.NoError(t, err)

		s, err := compileValue(result)
		assert.NoError(t, err)

		assert.Equal(t, testCases[i].Out, string(s))
	}
}
