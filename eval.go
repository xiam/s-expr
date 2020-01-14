package sexpr

import (
	"errors"
	"fmt"
	"log"
)

var (
	errFunctionClosed = errors.New("function is closed")
	errStreamClosed   = errors.New("stream is closed")
)

var (
	errUndefinedValue    = errors.New("undefined value")
	errUndefinedFunction = errors.New("undefined function")
	errClosedChannel     = errors.New("closed channel")
)

var defaultContext = NewContext(nil).Name("root")

func init() {
	defaultContext.executable = true
}

func Defn(name string, fn Function) {
	wrapper := func(ctx *Context) error {
		if err := fn(ctx); err != nil {
			ctx.Exit(err)
			return err
		}
		ctx.Exit(nil)
		return nil
	}
	if err := defaultContext.Set(name, NewFunctionValue(wrapper)); err != nil {
		log.Fatal("Defn: %w", err)
	}
}

func execArgument(ctx *Context, value *Value) (*Value, error) {
	switch value.Type {
	case ValueTypeFunction:
		newCtx := NewContext(ctx).Name("argument")
		go func() {
			defer newCtx.Exit(nil)
			if err := value.Function()(newCtx); err != nil {
				panic(err.Error())
			}
		}()
		col, err := newCtx.Collect()
		return col[0], err
	case ValueTypeList:
		panic("list")
	}
	return value, nil
}

func execFunctionBody(ctx *Context, body *Value) error {

	switch body.Type {
	case ValueTypeFunction:
		newCtx := NewClosure(ctx).Name("exec-body")
		go func() error {
			defer newCtx.Exit(nil)
			return body.Function()(newCtx)
		}()
		values, err := newCtx.Result()
		if err != nil {
			return err
		}
		ctx.Yield(values.List()...)
		return nil
	case ValueTypeList:
		for _, item := range body.List() {
			if err := execFunctionBody(ctx, item); err != nil {
				return err
			}
		}
		return nil
	default:
		panic("unhandled")
	}
}

func prepareFunc(values []*Value) *Value {
	return NewFunctionValue(Function(func(ctx *Context) error {

		fn := values[0]

		if len(values) == 1 {
			switch fn.Type {
			case ValueTypeInt, ValueTypeAtom, ValueTypeList:
				ctx.Yield(fn)
				return nil
			}
		}

		if fn.Type == ValueTypeFunction {
			var err error
			fn, err = execArgument(ctx, fn)
			if err != nil {
				return err
			}
		}

		switch fn.Type {
		case ValueTypeFunction, ValueTypeList, ValueTypeAtom, ValueTypeInt:
			ctx.Yield(fn)
			return nil
		}

		fnName := fn.raw()
		fn, err := ctx.Get(fnName)
		if err != nil {
			if err == errUndefinedFunction {
				log.Fatalf("undefined function %q", fnName)
				return fmt.Errorf("undefined function %q", fnName)
			}
			return err
		}

		//fnCtx := NewClosure(ctx).Executable()

		go func() {
			defer ctx.Close()

			for i := 1; i < len(values) && ctx.Accept(); i++ {
				//fnCtx.
				ctx.Push(values[i])
			}
		}()

		/*
			go func() {
				//defer fnCtx.Exit(nil)

			}()
		*/
		return fn.Function()(ctx)

		/*
			result, err := fnCtx.Result()
			if err != nil {
				log.Printf("err.res: %v", err)
				return err
			}
			for i := 0; i < len(result.List()); i++ {
				ctx.Yield(result.List()[i])
			}

			return nil
		*/

	}))
}

func evalContextList(ctx *Context, nodes []*Node) error {
	for i := range nodes {
		err := evalContext(ctx, nodes[i])
		if err != nil {
			//ctx.outErr <- err
			return err
		}
	}

	return nil
}

func evalContext(ctx *Context, node *Node) error {

	switch node.Type {
	case NodeTypeAtom:

		return ctx.Yield(&node.Value)

	case NodeTypeList:

		newCtx := NewContext(ctx).Name("list")
		go func() {
			defer newCtx.Exit(nil)
			err := evalContextList(newCtx, node.Children)
			if err != nil {
				return
			}
		}()

		value, err := newCtx.Result()
		if err != nil {
			return err
		}

		return ctx.Yield(value)

	case NodeTypeMap:

		newCtx := NewClosure(ctx).Name("map")
		go func() error {
			defer newCtx.Exit(nil)

			err := evalContextList(newCtx, node.Children)
			if err != nil {
				return err
			}
			return nil
		}()

		result := map[Value]*Value{}
		var key *Value
		for {
			value, err := newCtx.Output()
			if err != nil {
				if err == errClosedChannel {
					value, err := NewValue(result)
					if err != nil {
						return err
					}
					return ctx.Yield(value)
				}
				return err
			}
			if key == nil {
				key = value
				result[*key] = Nil
			} else {
				result[*key] = value
				key = nil
			}
		}

		panic("unreachable")

	case NodeTypeExpression:

		newCtx := NewClosure(ctx).Name("expr-eval").NonExecutable()
		go func() error {
			defer newCtx.Exit(nil)

			err := evalContextList(newCtx, node.Children)
			if err != nil {
				return err
			}

			return nil
		}()

		values, err := newCtx.Result()
		if err != nil {
			return err
		}

		fn := prepareFunc(values.List())

		if ctx.IsExecutable() {
			execCtx := NewContext(ctx).Name("expr-exec")

			go func() {
				defer execCtx.Exit(nil)

				if err := fn.Function()(execCtx); err != nil {
					log.Fatalf("ERR: %v", err)
				}
			}()

			values, err := execCtx.Result()
			if err != nil {
				return err
			}

			if len(values.List()) == 1 {
				return ctx.Yield(values.List()[0])
			}

			return ctx.Yield(values)
		}

		return ctx.Yield(fn)
	}

	panic("unreachable")
}

func eval(node *Node) (*Context, interface{}, error) {
	newCtx := NewContext(defaultContext).Name("eval")

	go func() {
		defer newCtx.Exit(nil)

		if err := evalContext(newCtx, node); err != nil {
			log.Fatalf("EVAL.CONTEXT: %v", err)
			return
		}
	}()

	values, err := newCtx.Collect()
	if err != nil {
		return nil, nil, err
	}

	if len(values) == 0 {
		return newCtx, nil, nil
	}

	if len(values) == 1 {
		return newCtx, values[0], nil
	}

	return newCtx, values, nil
}
