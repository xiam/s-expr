package sexpr

import (
	"errors"
	//"fmt"
	"log"
)

type contextState uint8

const (
	contextStateReady contextState = iota

	contextStateWaitForInput
	contextStateInput
	contextStateCloseInput
	contextStateOutput
	contextStateCloseOutput

	contextStateError
	contextStateDone
)

type contextMessage struct {
	state   contextState
	payload interface{}
}

var (
	errUndefinedValue    = errors.New("undefined value")
	errUndefinedFunction = errors.New("undefined function")
	errClosedChannel     = errors.New("closed channel")
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

func (st *symbolTable) Set(name string, value *Value) error {
	if st.t != symbolTableTypeDict {
		return errors.New("not a dictionary")
	}
	st.n[name] = &symbolTable{
		t: symbolTableTypeValue,
		v: value,
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

type Function func(*Context) (*Value, error)

type Context struct {
	Parent *Context

	in       chan Value
	inErr    chan error
	inClosed bool

	out       chan Value
	outErr    chan error
	outClosed bool

	message chan contextMessage

	lastArgument *Value

	done bool

	st *symbolTable
}

func (ctx *Context) closeIn() {
	close(ctx.in)
}

func (ctx *Context) closeOut() {
	close(ctx.out)
}

func (ctx *Context) exit(err error) error {
	return nil
}

func (ctx *Context) run() error {
	for {
		switch m := <-ctx.message; m.state {
		case contextStateInput:
			ctx.in <- m.payload.(Value)
		case contextStateCloseInput:
			ctx.closeIn()
		case contextStateOutput:
			ctx.out <- m.payload.(Value)
		case contextStateCloseOutput:
			ctx.closeOut()
		case contextStateDone:
			ctx.exit(m.payload.(error))
			return nil
		}
	}

	panic("unreachable")
}

func (ctx *Context) Next() bool {
	arg, ok := <-ctx.in
	if !ok {
		return false
	}
	ctx.lastArgument = &arg
	return true
}

func (ctx *Context) Arguments() ([]*Value, error) {
	if ctx.inClosed {
		return nil, errors.New("channel is closed")
	}
	args := []*Value{}
	for ctx.Next() {
		args = append(args, ctx.Argument())
	}
	return args, nil
}

func (ctx *Context) Argument() *Value {
	return ctx.lastArgument
}

func (ctx *Context) Exit(err error) {
	close(ctx.out)
	ctx.outClosed = true
	if err != nil {
		ctx.outErr <- err
	}
}

func (ctx *Context) Close() {
	if !ctx.inClosed {
		close(ctx.in)
	}
	ctx.inClosed = true
}

func (ctx *Context) Push(value *Value) error {
	if ctx.inClosed {
		return errors.New("channel is closed")
	}
	ctx.in <- *value
	return nil
}

func (ctx *Context) Return(values ...*Value) error {
	for i := range values {
		err := ctx.Yield(values[i])
		if err != nil {
			return err
		}
	}
	ctx.Exit(nil)
	return nil
}

func (ctx *Context) Yield(value *Value) error {
	if ctx.outClosed {
		log.Printf("YIELD: closed")
		return nil
	}
	if value == nil {
		panic("can't yield nil value")
	}
	log.Printf("YIELD: %v", value)
	ctx.out <- *value
	log.Printf("YIELDED: %v", value)
	return nil
}

func (ctx *Context) NextInput() (Value, error) {
	select {
	case err := <-ctx.inErr:
		return *Nil, err
	default:
		in, ok := <-ctx.in
		if !ok {
			return *Nil, errClosedChannel
		}
		log.Printf("NEXTINPUT: %v - %v", in, ok)
		//if in == nil {
		//	return Nil, nil
		//}
		return in, nil
	}

	panic("unreachable")
}

func (ctx *Context) NextOutput() (Value, error) {
	select {
	case err := <-ctx.outErr:
		return *Nil, err
	default:
		out, ok := <-ctx.out
		if !ok {
			return *Nil, errClosedChannel
		}
		log.Printf("NEXTOUTPUT: %v - %v", out, ok)
		//if out == nil {
		//	return Nil, nil
		//}
		return out, nil
	}

	panic("unreachable")
}

func (ctx *Context) Collect() (*Value, error) {
	values := []*Value{}

	for {
		select {
		case err := <-ctx.outErr:
			return nil, err
		default:
			out, ok := <-ctx.out
			log.Printf("inside collect: %v - %v", out, ok)
			if !ok {
				if len(values) == 1 {
					return values[0], nil
				}
				v, err := NewValue(values)
				log.Printf("inside collect 2: %v - %v", out, ok)
				return v, err
			}
			values = append(values, &out)
		}
	}

	panic("unreachable")
}

func (ctx *Context) Set(name string, value *Value) error {
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
	ctx := &Context{
		Parent: parent,

		message: make(chan contextMessage),

		in:    make(chan Value),
		inErr: make(chan error),

		out:    make(chan Value),
		outErr: make(chan error),

		st: &symbolTable{
			t: symbolTableTypeDict,
			n: make(map[string]*symbolTable),
		},
	}

	go func() {
		err := ctx.run()
		log.Printf("ctx: %v", err)
	}()

	return ctx
}

var defaultContext = NewContext(nil)

func RegisterPrefix(name string, fn Function) {
	value, err := NewValue(fn)
	if err != nil {
		log.Fatal("RegisterPrefix: %w", err)
	}

	if err := defaultContext.Set(name, value); err != nil {
		log.Fatal("RegisterPrefix: %w", err)
	}
}

func evalContextMap(ctx *Context, nodes []*Node) error {
	log.Printf("evalContextList")

	go func() {
		defer ctx.Close()

		for i := range nodes {
			err := evalContext(ctx, nodes[i])
			if err != nil {
				log.Printf("eval error: %v", err)
				ctx.outErr <- err
				return
			}
		}
	}()

	return nil
}

func evalContextList(ctx *Context, nodes []*Node) error {
	log.Printf("evalContextList")

	go func() {
		defer ctx.Close()

		for i := range nodes {
			err := evalContext(ctx, nodes[i])
			if err != nil {
				log.Printf("eval error: %v", err)
				ctx.outErr <- err
				return
			}
		}
	}()

	return nil
}

func evalContext(ctx *Context, node *Node) error {

	log.Printf("evalContext: %v", node)

	switch node.Type {
	case NodeTypeAtom:
		log.Printf("ATOM: %v", node)

		return ctx.Yield(&node.Value)

	case NodeTypeList:
		log.Printf("LIST: %v", node)

		newCtx := NewContext(ctx)
		err := evalContextList(newCtx, node.Children)
		if err != nil {
			return err
		}

		value, err := newCtx.Collect()
		if err != nil {
			return err
		}

		return ctx.Yield(value)

	case NodeTypeMap:
		log.Printf("MAP: %#v", node)

		newCtx := NewContext(ctx)
		err := evalContextList(newCtx, node.Children)
		if err != nil {
			return err
		}

		result := map[Value]*Value{}
		var key *Value
		for {
			value, err := newCtx.NextOutput()
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
				key = &value
				result[*key] = Nil
			} else {
				result[*key] = &value
				key = nil
			}
		}

		panic("unreachable")

	case NodeTypeExpression:
		log.Printf("EXPRESSION: %v", node)

		newCtx := NewContext(ctx)

		err := evalContextList(newCtx, node.Children)
		if err != nil {
			return err
		}

		collected := make(chan *Value)

		result := []*Value{}

		var expr *Value
		var fnCtx *Context

		for {
			log.Printf("WAIT FOR NEXT OUTOUT")
			value, err := newCtx.NextOutput()
			log.Printf("GOT NEXT: %v - %v", value, err)

			if err != nil {
				if err == errClosedChannel {
					if expr != nil && len(result) < 1 {
						switch expr.Type {
						case ValueTypeAtom, ValueTypeNil, ValueTypeInt, ValueTypeFloat, ValueTypeList, ValueTypeMap, ValueTypeBinary, ValueTypeBool:
							log.Printf("YIELD")
							return ctx.Yield(expr)
						}
						return errors.New("unsupported function - 1")
					}
					if fnCtx == nil {
						return ctx.Yield(Nil)
					}
					fnCtx.Close()
					return ctx.Yield(<-collected)
				}
				return err
			}

			if fnCtx != nil {
				log.Printf("PUCHING")
				fnCtx.Push(&value)
				continue
			}

			if expr != nil {
				return errors.New("unsupported function - 2")
			}

			fnCtx = NewContext(ctx)
			switch value.Type {
			case ValueTypeString:
				fnVal, err := ctx.Get(value.String())
				if err != nil {
					return err
				}
				fn := fnVal.Function()
				go func() {
					out, err := fn(fnCtx)
					log.Printf("OUT.2: %v, ERR: %v", out, err)
					if err != nil {
						fnCtx.outErr <- err
					}
				}()
				go func() {
					value, err := fnCtx.Collect()
					log.Printf("[COLLECTED] OUT.1: %v, ERR: %v", value, err)
					collected <- value
					log.Printf("DONE COLLECTING")
				}()
			default:
				expr = &value
			}
		}

		panic("unreachable")

	}

	return nil
}

func eval(node *Node) (*Context, error) {
	newCtx := NewContext(defaultContext)

	go func() {
		if err := evalContext(newCtx, node); err != nil {
			log.Printf("ERR: %v", err)
			return
		}
	}()

	log.Printf("READ")
	value := <-newCtx.out
	close(newCtx.out)
	log.Printf("E.ARGS: %v", value)
	return newCtx, nil
}
