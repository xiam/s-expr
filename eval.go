package sexpr

import (
	"errors"
	//"fmt"
	"log"
)

var (
	errUndefinedValue    = errors.New("undefined value")
	errUndefinedFunction = errors.New("undefined function")
)

type symbolTableType uint8

const (
	symbolTableTypeValue symbolTableType = iota
	symbolTableTypeDict
)

type symbolTable struct {
	t symbolTableType
	n map[string]*symbolTable
	v *Value
}

func (st *symbolTable) Set(name string, value interface{}) error {
	if st.t != symbolTableTypeDict {
		return errors.New("not a dictionary")
	}
	v, err := NewValue(value)
	if err != nil {
		return err
	}
	st.n[name] = &symbolTable{t: symbolTableTypeValue, v: v}
	if err != nil {
		return err
	}
	return nil
}

func (st *symbolTable) Get(name string) (*Value, error) {
	if st.t != symbolTableTypeDict {
		return nil, errors.New("not a dictionary")
	}
	if v, ok := st.n[name]; ok {
		if v.t == symbolTableTypeValue {
			return v.v, nil
		}
		return nil, errors.New("key is not a value")
	}
	return nil, errors.New("no such key")
}

type Function func(*Context) (Value, error)

type Context struct {
	Parent *Context

	in  chan *Value
	out chan *Value

	st *symbolTable
}

func (ctx *Context) Set(name string, value interface{}) error {
	return ctx.st.Set(name, value)
}

func (ctx *Context) Get(name string) (*Value, error) {
	fn, err := ctx.st.Get(name)
	if err != nil {
		if ctx.Parent == nil {
			return nil, errUndefinedFunction
		}
		return ctx.Parent.Get(name)
	}
	return fn, nil
}

func NewContext(parent *Context) *Context {
	return &Context{
		Parent: parent,
		in:     make(chan *Value),
		out:    make(chan *Value),
		st: &symbolTable{
			t: symbolTableTypeDict,
			n: make(map[string]*symbolTable),
		},
	}
}

var defaultContext = NewContext(nil)

func RegisterPrefix(name string, fn Function) error {
	return defaultContext.Set(name, fn)
}

func evalContext(ctx *Context, node *Node) error {

	log.Printf("ui-node: %v", node)

	switch node.Type {
	case NodeTypeAtom:
		log.Printf("ATOM: %#v", node.Value)
		ctx.out <- &node.Value

	case NodeTypeList:
		log.Printf("LIST: %v", node)

		newCtx := NewContext(ctx)
		result := make(chan []Value)
		go func() {
			values := []Value{}
			for value := range newCtx.out {
				log.Printf("value: %v", value)
				values = append(values, *value)
			}
			result <- values
		}()

		for i := range node.Children {
			err := evalContext(newCtx, node.Children[i])
			if err != nil {
				log.Printf("eval error: %v", err)
				break
			}
		}
		close(newCtx.out)

		value, err := NewValue(<-result)
		if err != nil {
			return err
		}

		ctx.out <- value
		return nil

	case NodeTypeExpression:
		log.Printf("EXPRESSION: %v", node)

		newCtx := NewContext(ctx)

		result := make(chan []Value)
		go func() {
			values := []Value{}
			for value := range newCtx.out {
				log.Printf("value: %v", value)
				values = append(values, *value)
			}
			result <- values
		}()

		for i := range node.Children {
			err := evalContext(newCtx, node.Children[i])
			if err != nil {
				log.Printf("eval error: %v", err)
				break
			}
		}
		close(newCtx.out)

		value, err := NewValue(<-result)
		if err != nil {
			return err
		}

		ctx.out <- value
		return nil
	}

	/*
		log.Printf("EVAL-> %v", node)
		node.Serve()
		log.Printf("EVAL-1")

		switch node.Type {
		case NodeTypeAtom:
			log.Printf("ATOM: %v", node)

		case NodeTypeList:
			newCtx := NewContext(ctx)

			for {
				nextNode := node.Next()
				if nextNode == nil {
					break
				}
				if err := evalContext(newCtx, nextNode); err != nil {
					return err
				}
			}

		case NodeTypeExpression:
			log.Printf("EXPRE")
			for {
				nextNode := node.Next()
				log.Printf("NEXT: %v", nextNode)
				if nextNode == nil {
					break
				}
				if err := evalContext(ctx, nextNode); err != nil {
					return err
				}
			}
		default:
			log.Printf("BREAK")
		}

		/*

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
	*/

	return nil
}

func eval(node *Node) (*Context, error) {
	newCtx := NewContext(defaultContext)

	go func() {
		defer close(newCtx.out)
		if err := evalContext(newCtx, node); err != nil {
			log.Printf("ERR: %v", err)
			return
		}
	}()

	value := <-newCtx.out
	log.Printf("E.ARGS: %v", value)
	return newCtx, nil
}
