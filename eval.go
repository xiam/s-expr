package sexpr

import (
	"errors"
	"fmt"
	"log"
)

var (
	errUndefinedValue    = errors.New("undefined value")
	errUndefinedFunction = errors.New("undefined function")
)

type Function func(*Context) (interface{}, error)

type Context struct {
	Parent *Context
	args   chan interface{}

	prefixMap map[string]Function
	envMap    map[string]interface{}
}

func (e *Context) pushArgument(arg interface{}) {
	log.Printf("#> %v", arg)
	e.args <- arg
}

func (e *Context) Close() {
	close(e.args)
}

func (e *Context) Argument() (interface{}, error) {
	log.Printf("moar...")
	arg, ok := <-e.args
	if !ok {
		log.Printf("#< [!]")
		return nil, errors.New("no such argument")
	}
	log.Printf("#< %v", arg)
	return arg, nil
}

func (e *Context) Arguments() ([]interface{}, error) {
	args := []interface{}{}
	for {
		arg, err := e.Argument()
		if err != nil {
			break
		}
		args = append(args, arg)
	}
	return args, nil
}

func (e *Context) SetFn(name string, value Function) error {
	e.prefixMap[name] = value
	return nil
}

func (e *Context) GetFn(name string) (Function, error) {
	if val, ok := e.prefixMap[name]; ok {
		return val, nil
	}
	if e.Parent == nil {
		return nil, errUndefinedFunction
	}
	return e.Parent.GetFn(name)
}

func (e *Context) Set(name string, value interface{}) error {
	e.envMap[name] = value
	return nil
}

func (e *Context) Get(name string) (interface{}, error) {
	if value, ok := e.envMap[name]; ok {
		return value, nil
	}
	if e.Parent == nil {
		return nil, errUndefinedValue
	}
	return e.Parent.Get(name)
}

var (
	prefixMap = map[string]Function{}
)

func NewContext(parent *Context) *Context {
	return &Context{
		Parent:    parent,
		args:      make(chan interface{}),
		prefixMap: make(map[string]Function),
		envMap:    make(map[string]interface{}),
	}
}

var defaultContext = NewContext(nil)

func RegisterPrefix(name string, fn Function) {
	defaultContext.SetFn(name, fn)
}

func evalContext(ctx *Context, node *Node) error {
	log.Printf("EVAL-> %v", node)

	switch node.Type {

	case NodeTypeAtom:
		log.Printf("PUSHED ARGUMENT")
		ctx.pushArgument(node.Value)

	case NodeTypeArray:
		newCtx := NewContext(ctx)

		go func() {
			defer newCtx.Close()
			for i := range node.Children {
				err := evalContext(newCtx, node.Children[i])
				if err != nil {
					log.Printf("eval error: %v", err)
					break
				}
			}
		}()

		args, err := newCtx.Arguments()
		if err != nil {
			log.Printf("eval error: %v", err)
			return err
		}
		log.Printf("READ ALL ARGUMENTS")
		ctx.pushArgument(args)

	case NodeTypeList:
		newCtx := NewContext(ctx)

		go func() {
			defer newCtx.Close()
			for i := range node.Children {
				err := evalContext(newCtx, node.Children[i])
				if err != nil {
					log.Printf("eval error: %v", err)
					break
				}
			}
		}()

		firstAtom, err := newCtx.Argument()
		if err != nil {
			return err
		}

		switch expr := firstAtom.(type) {
		case string:
			log.Printf("LOOKUP: %v", expr)
			fn, err := newCtx.GetFn(expr)
			if err == nil {
				value, err := fn(newCtx)
				log.Printf("FN: %v, %v", value, err)
				if err != nil {
					return err
				}
				ctx.pushArgument(value)
				return nil
			}
			return fmt.Errorf("undefined expression %v", expr)
		case bool:
			if expr {
				args, err := newCtx.Arguments()
				if err != nil {
					log.Printf("eval error: %v", err)
					return nil
				}
				ctx.pushArgument(append([]interface{}{firstAtom}, args...))
			} else {
				// drain arguments
				newCtx.Arguments()
				log.Printf("drained arguments")
			}
		default:
			args, err := newCtx.Arguments()
			if err != nil {
				log.Printf("eval error: %v", err)
				return nil
			}
			ctx.pushArgument(append([]interface{}{firstAtom}, args...))
		}

	default:
		panic("unknown type")
	}

	return nil
}

func eval(node *Node) (*Context, error) {
	newCtx := NewContext(defaultContext)

	go func() {
		defer newCtx.Close()

		if err := evalContext(newCtx, node); err != nil {
			log.Printf("ERR: %v", err)
			return
		}
	}()

	args, err := newCtx.Arguments()
	log.Printf("E.ARGS: %v, E.ERR: %v", args, err)

	return newCtx, nil
}
