package sexpr

import (
	"fmt"
	"log"
	//"strings"
	"errors"
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
			In:  `1 2 3`,
			Out: `[1 2 3]`,
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
			In:  `(((1)))`,
			Out: `[1]`,
		},
		{
			In:  `([1])`,
			Out: `[[1]]`,
		},
		{
			In:  `([[1]])`,
			Out: `[[[1]]]`,
		},
		{
			In:  `[([1])]`,
			Out: `[[[1]]]`,
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
			In:  `[(nop [ [ (echo :hello) ]])]`,
			Out: `[[:nil]]`,
		},
		{
			In:  `[(print "hello " "world!")]`,
			Out: `[[:nil]]`,
		},
		{
			In:  `(echo "foo" "bar")`,
			Out: `[["foo" "bar"]]`,
		},
		{
			In:  `(["foo" "bar"])`,
			Out: `[["foo" "bar"]]`,
		},
		{
			In:  `([["foo" "bar"]])`,
			Out: `[[["foo" "bar"]]]`,
		},
		{
			In:  `((([["foo" "bar"]])))`,
			Out: `[[["foo" "bar"]]]`,
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
			In:  `(10)`,
			Out: `[10]`,
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
		{
			In:  `(set foo 1)`,
			Out: `[1]`,
		},
		{
			In:  `(get foo)`,
			Out: `[:nil]`,
		},
		{
			In:  `(get foo) (set foo 3) (get foo) (get foo)`,
			Out: `[:nil 3 3 3]`,
		},
		{
			In:  `(echo (set foo 1) (get foo))`,
			Out: `[[1 1]]`,
		},
		{
			In:  `(echo "hello" "world!")`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(echo "hello" (echo "world!"))`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(echo "hello" (echo (echo (echo "world!"))))`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(:true)`,
			Out: `[:true]`,
		},
		{
			In:  `("anyvalue")`,
			Out: `["anyvalue"]`,
		},
		{
			In:  `(  123 )`,
			Out: `[123]`,
		},
		{
			In:  `(:true :true)`,
			Out: `[:true]`,
		},
		{
			In:  `(:true :false :true :true :false)`,
			Out: `[:true]`,
		},
		{
			In:  `(:false)`,
			Out: `[:false]`,
		},
		{
			In:  `(:false :true :true)`,
			Out: `[:false]`,
		},
		{
			In:  `(:true "hello")`,
			Out: `[:true]`,
		},
		{
			In:  `(:true (echo "hello" (echo "world")))`,
			Out: `[:true]`,
		},
		{
			In:  `(:false (echo "hello" (echo "world")))`,
			Out: `[:false]`,
		},
		{
			In:  `(:true (echo "hello" "world!"))`,
			Out: `[:true]`,
		},
		{
			In:  `(echo "hello" (echo (echo (echo "world!"))))`,
			Out: `[["hello" "world!"]]`,
		},
		{
			In:  `(:true (echo "hello" (echo (echo (echo "world!")))))`,
			Out: `[:true]`,
		},
		{
			In:  `(:false (echo "hello" "world!"))`,
			Out: `[:false]`,
		},
		{
			In:  `(defn foo [word] (echo (get word)))`,
			Out: `[:true]`,
		},
		{
			In:  `(defn foo [word] (echo (get word))) (foo "HEY")`,
			Out: `[:true "HEY"]`,
		},
	}

	Defn("+", func(ctx *Context) error {
		result := int64(0)
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			result += value.Int()
		}
		ctx.Yield(NewIntValue(result))
		return nil
	})

	Defn(":false", func(ctx *Context) error {
		ctx.Yield(False)
		return nil
	})

	Defn(":true", func(ctx *Context) error {
		for ctx.Next() {
			_, err := ctx.Argument()
			if err != nil {
				return err
			}
		}

		ctx.Yield(True)
		return nil
	})

	Defn("echo", func(ctx *Context) error {
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			ctx.Yield(value)
		}

		return nil
	})

	Defn("=", func(ctx *Context) error {
		var first *Value
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}

			if first == nil {
				first = value
				continue
			}

			if (*first).raw() != value.raw() {
				ctx.Yield(False)
				return nil
			}
		}

		ctx.Yield(True)
		return nil
	})

	Defn("nop", func(ctx *Context) error {
		ctx.Yield(Nil)

		return nil
	})

	Defn("defn", func(ctx *Context) error {
		var name, params, fn *Value

		ctx = ctx.NonExecutable()
		for i := 0; ctx.Next(); i++ {
			arg, err := ctx.Argument()
			if err != nil {
				return err
			}

			switch i {
			case 0:
				name = arg
			case 1:
				params = arg
			default:
				fn = arg
			}
		}

		paramsList := params.List()

		wrapperFn := NewFunctionValue(Function(func(ctx *Context) error {
			for i := 0; ctx.Next() && i < len(paramsList); i++ {
				arg, err := ctx.Argument()
				if err != nil {
					return err
				}
				ctx.Set(paramsList[i].raw(), arg)
			}
			return fn.Function()(ctx)
		}))

		ctx.Set(name.raw(), wrapperFn)

		ctx.Yield(True)
		return nil
	})

	Defn("print", func(ctx *Context) error {
		for ctx.Next() {
			value, err := ctx.Argument()
			if err != nil {
				return err
			}
			fmt.Printf("%s", value.raw())
		}

		ctx.Yield(Nil)
		return nil
	})

	Defn("get", func(ctx *Context) error {
		var name *Value
		for i := 0; ctx.Next(); i++ {
			argument, err := ctx.Argument()
			if err != nil {
				return err
			}
			switch i {
			case 0:
				name = argument
			default:
				return errors.New("expecting one argument")
			}
		}

		value, err := ctx.Get(name.raw())
		if err != nil {
			ctx.Yield(Nil)
			return nil
		}

		ctx.Yield(value)
		return nil
	})

	Defn("set", func(ctx *Context) error {
		var name, value *Value
		for i := 0; ctx.Next(); i++ {
			argument, err := ctx.Argument()
			if err != nil {
				return err
			}
			switch i {
			case 0:
				name = argument
			case 1:
				value = argument
			default:
				return errors.New("expecting two arguments")
			}
		}

		ctx.Set(name.raw(), value)
		ctx.Yield(value)

		return nil
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
