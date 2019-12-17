package sexpr

import (
	"errors"
	//"fmt"
	"log"
	"sync"
)

var (
	errFunctionClosed = errors.New("function is closed")
	errStreamClosed   = errors.New("stream is closed")
)

type contextState uint8

const (
	contextStateReady contextState = iota

	contextStateAccept
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

	ticket chan struct{}

	inMu  sync.Mutex
	mu    sync.Mutex
	argMu sync.Mutex

	in       chan *Value
	inClosed bool

	doneAccept chan struct{}

	out       chan *Value
	outClosed bool

	accept chan struct{}

	exitStatus   error
	lastArgument *Value

	st *symbolTable
}

func (ctx *Context) closeIn() {
	if ctx.inClosed {
		return
	}

	close(ctx.accept)
	close(ctx.in)

	ctx.inClosed = true
}

func (ctx *Context) exit(err error) error {
	if err != nil {
		ctx.exitStatus = err
	}
	return nil
}

func (ctx *Context) Next() bool {
	ctx.mu.Lock()
	if ctx.inClosed {
		ctx.mu.Unlock()
		return false
	}
	ctx.accept <- struct{}{}
	ctx.mu.Unlock()

	var ok bool
	ctx.lastArgument, ok = <-ctx.in
	if !ok {
		return false
	}
	return true
}

func (ctx *Context) Arguments() ([]*Value, error) {
	if ctx.inClosed {
		return nil, errStreamClosed
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
	if ctx.outClosed {
		return
	}

	close(ctx.out)
	ctx.outClosed = true
	ctx.Close()
}

func (ctx *Context) Close() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.inClosed {
		return
	}
	ctx.inClosed = true
	close(ctx.doneAccept)
	close(ctx.accept)
	close(ctx.in)
}

func (ctx *Context) Push(value *Value) error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.inClosed {
		return errors.New("channel is closed")
	}
	ctx.in <- value
	return nil
}

func (ctx *Context) Return(values ...*Value) error {
	defer ctx.Exit(nil)

	for i := range values {
		err := ctx.Yield(values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *Context) Accept() bool {
	if ctx.inClosed {
		return false
	}
	select {
	case <-ctx.doneAccept:
	case <-ctx.accept:
		return true
	}
	return false
}

func (ctx *Context) Yield(value *Value) error {
	if ctx.outClosed {
		return nil
	}
	if value == nil {
		panic("can't yield nil value")
	}
	ctx.out <- value
	return nil
}

func (ctx *Context) Output() (*Value, error) {
	out, ok := <-ctx.out
	if !ok {
		return nil, errClosedChannel
	}
	return out, nil
}

func (ctx *Context) Result() (*Value, error) {
	values, err := ctx.Collect()
	if err != nil {
		return nil, err
	}
	return NewValue(values)
}

func (ctx *Context) Collect() ([]*Value, error) {
	values := []*Value{}

	for {
		value, err := ctx.Output()
		if err == errClosedChannel {
			break
		}
		values = append(values, value)
	}

	return values, nil
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
		ticket: make(chan struct{}),

		accept:     make(chan struct{}, 1),
		doneAccept: make(chan struct{}),

		in:  make(chan *Value),
		out: make(chan *Value),

		st: &symbolTable{
			t: symbolTableTypeDict,
			n: make(map[string]*symbolTable),
		},
	}

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
	go func() {
		defer ctx.Exit(nil)

		for i := range nodes {
			err := evalContext(ctx, nodes[i])
			if err != nil {
				//ctx.outErr <- err
				return
			}
		}
	}()

	return nil
}

func evalContextList(ctx *Context, nodes []*Node) error {
	go func() {
		defer ctx.Exit(nil)

		for i := range nodes {
			err := evalContext(ctx, nodes[i])
			if err != nil {
				//ctx.outErr <- err
				return
			}
		}
	}()

	return nil
}

func evalContext(ctx *Context, node *Node) error {
	switch node.Type {
	case NodeTypeAtom:
		return ctx.Yield(&node.Value)

	case NodeTypeList:
		newCtx := NewContext(ctx)
		err := evalContextList(newCtx, node.Children)
		if err != nil {
			return err
		}

		value, err := newCtx.Result()
		if err != nil {
			return err
		}

		return ctx.Yield(value)

	case NodeTypeMap:
		newCtx := NewContext(ctx)
		err := evalContextList(newCtx, node.Children)
		if err != nil {
			return err
		}

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
		newCtx := NewContext(ctx)
		err := evalContextList(newCtx, node.Children)
		if err != nil {
			return err
		}

		collected := make(chan []*Value)

		var expr *Value
		var fnCtx *Context

		for {
			value, err := newCtx.Output()

			if err != nil {
				if err == errClosedChannel {
					if expr != nil {
						switch expr.Type {
						case ValueTypeAtom, ValueTypeNil, ValueTypeInt, ValueTypeFloat, ValueTypeList, ValueTypeMap, ValueTypeBinary, ValueTypeBool:
							return ctx.Yield(expr)
						}
						return errors.New("unsupported function - 1")
					}
					if fnCtx == nil {
						return ctx.Yield(Nil)
					}
					fnCtx.Close()
					values := <-collected

					if len(values) == 0 {
						ctx.Yield(Nil)
						return nil
					}
					if len(values) == 1 {
						ctx.Yield(values[0])
						return nil
					}

					value, err := NewValue(values)
					if err != nil {
						return err
					}
					ctx.Yield(value)

					return nil
				}
				return err
			}

			if fnCtx != nil {
				accept := fnCtx.Accept()
				if accept {
					fnCtx.Push(value)
				}
				continue
			}

			if expr != nil {
				return errors.New("unsupported function - 2")
			}

			fnCtx = NewContext(ctx)
			switch value.Type {
			case ValueTypeString:
				fnVal, err := ctx.Get(value.raw())
				if err != nil {
					return err
				}
				fn := fnVal.Function()
				go func() {
					out, err := fn(fnCtx)
					if err != nil {
						//fnCtx.outErr <- err
					}
					_ = out
				}()
				go func() {
					values, _ := fnCtx.Collect()
					collected <- values
				}()
			default:
				expr = value
			}
		}

		panic("unreachable")

	}

	return nil
}

func eval(node *Node) (*Context, interface{}, error) {
	newCtx := NewContext(defaultContext)

	go func() {
		defer newCtx.Exit(nil)

		if err := evalContext(newCtx, node); err != nil {
			return
		}
	}()

	values, err := newCtx.Collect()
	if err != nil {
		return nil, nil, err
	}

	return newCtx, values, nil
}
