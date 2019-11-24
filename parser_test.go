package sexpr

import (
	//"errors"
	//"fmt"
	"log"
	//"strings"
	"testing"
	//"time"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
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
	}

	for i := range testCases {
		root, err := parse([]byte(testCases[i].In))
		assert.NoError(t, err)
		assert.NotNil(t, root)

		printNode(root)
		s := compileNode(root)
		log.Printf("compiled: %v", string(s))

		log.Printf("TREE: %#v", root)
	}

}

func SkipTestParser(t *testing.T) {

	//tree, err := parse([]byte(`[] [1 [2 3] [4 5 6 [7 "[" 8 9]]] (fn zero [(set cero 0) (print (get cero))]) (zero) (set anakin [1 2 [3 4] 5]) (set foo 69) (set bar (get foo)) (get foo bar anakin) (+ "1234 567" (- 56 "6" "" 6 (7)) 9 78) (+ 7 7 7 7) () (get foo) (+ 2 3 (+ 5 5)) (1 2 3 4)`))

	//tree, err := parse([]byte(`[[123 9 8] 45 6] [1 4 5 655]`))
	tree, err := parse([]byte(`
	[1 2 3 [4 5 [6 7 8]]]
	[foo bar baz]
	`))
	/*
		tree, err := parse([]byte(`

		[1 2 3 [4 5 [6 7 8]]]

		(when
			((= 1 0) 11)
			((= 1 1) 99 [110 [121]])
			((= 2 1) 33)
		)`))
	*/

	/*
		tree, err := parse([]byte(`
			[7 8 1]

			(fn zero []
			[
				(set cero (now))
				(set uno "UNO")
				(set dos "DOS")
				(print (get cero))
				(print (get uno))
			])

			(zero)
			(zero)
			(zero)
			(zero)

			(fn WAKA [n] [
				(print (get n))
			])

			(WAKA 6)

			(fn FIB [n] [
				(when
					((= 1 1) 55)
					((= (get n) 0) 0)
					((= (get n) 1) 1)
					# (+ (FIB (- (get n) 1) (- (get n) 2)))
				)
			])

			(FIB 5)

			`))
	*/
	assert.NoError(t, err)

	/*
		RegisterPrefix("set", func(ctx *Context) (interface{}, error) {
			args, err := ctx.Arguments()
			if err != nil {
				return nil, err
			}

			if len(args) < 2 {
				return false, errors.New("set requires two arguments")
			}
			var value interface{}
			if len(args) == 2 {
				value = args[1]
			} else {
				value = args[1:]
			}
			if err := ctx.Set(fmt.Sprintf("%v", args[0]), value); err != nil {
				return false, err
			}
			return true, nil
		})

		RegisterPrefix("now", func(ctx *Context) (interface{}, error) {
			return fmt.Sprintf("%v", time.Now()), nil
		})

		RegisterPrefix("get", func(ctx *Context) (interface{}, error) {
			args, err := ctx.Arguments()
			if err != nil {
				return nil, err
			}

			if len(args) < 1 {
				return nil, errors.New("get requires one argument")
			}
			values := []interface{}{}
			for i := range args {
				value, err := ctx.Get(fmt.Sprintf("%v", args[i]))
				log.Printf("WAT")
				if err != nil {
					return nil, err
				}
				values = append(values, value)
			}
			if len(values) == 1 {
				return values[0], nil
			}
			return values, nil
		})

		RegisterPrefix("fn", func(ctx *Context) (interface{}, error) {
			args, err := ctx.Arguments()
			if err != nil {
				return nil, err
			}

			if len(args) < 2 {
				return false, errors.New("fn requires one argument")
			}

			env := []string{}
			argset, ok := args[1].([]interface{})
			if ok {
				for i := range argset {
					env = append(env, argset[i].(string))
				}
			}
			log.Printf("ENV: %v", env)

			sargs := []string{}
			body := args[2].([]interface{})
			for i := range body {
				sargs = append(sargs, body[i].(string))
			}
			content := strings.Join(sargs, " ")
			log.Printf("EMEDDED: --%#v--\n", content)

			tree, err := parse([]byte(content))
			if err != nil {
				return false, fmt.Errorf("parse error: %v", err)
			}

			ctx.SetFn(args[0].(string), func(ctx *Context) (interface{}, error) {
				args, err := ctx.Arguments()
				if err != nil {
					return nil, err
				}

				newCtx := NewContext(ctx)
				defer newCtx.Close()

				log.Printf("ARGS: %v", args)
				for i := range env {
					if i < len(args) {
						newCtx.Set(env[i], args[i])
					}
				}

				if err := evalContext(newCtx, tree); err != nil {
					return false, err
				}
				return true, nil
			})

			return true, nil
		})

		RegisterPrefix("=", func(ctx *Context) (interface{}, error) {
			args, err := ctx.Arguments()
			if err != nil {
				return nil, err
			}

			left := args[0].(interface{})
			right := args[1].(interface{})

			log.Printf("LEFT: %v, RIGHT: %v", left, right)
			if fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right) {
				log.Printf("EQ TRUE")
				return true, nil
			}

			return false, nil
		})

		RegisterPrefix("when", func(ctx *Context) (interface{}, error) {
			log.Printf("WHEN!")
			for {
				log.Printf("WHEN WAIT")
				arg, err := ctx.Argument()
				if err != nil {
					return false, err
				}

				if arg.([]interface{})[0].(bool) == true {
					return arg.([]interface{})[1:], nil
				}
				return false, nil
			}
			return false, nil
		})

		RegisterPrefix("print", func(ctx *Context) (interface{}, error) {
			args, err := ctx.Arguments()
			if err != nil {
				return nil, err
			}
			fmt.Printf("PRINT %#v\n", args)
			return true, nil
		})

		RegisterPrefix("+", func(ctx *Context) (interface{}, error) {
			args, err := ctx.Arguments()
			if err != nil {
				return nil, err
			}
			return len(args), nil
		})

		RegisterPrefix("-", func(ctx *Context) (interface{}, error) {
			args, err := ctx.Arguments()
			if err != nil {
				return nil, err
			}
			return len(args), nil
		})
	*/

	{
		ctx, err := eval(tree)
		assert.NoError(t, err)

		log.Printf("CTX: %#v", ctx)
	}
}
