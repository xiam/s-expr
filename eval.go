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
	/*
		if len(values) < 1 {
			return Nil, nil
		}
	*/
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

func NewClosure(parent *Context) *Context {
	ctx := NewContext(parent)
	ctx.st = parent.st
	return ctx
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

type exprContext struct {
	fn  *Value
	ctx *Context
}

func evalContextExpression(ctx *Context, nodes []*Node) error {
	fnCtx := NewClosure(ctx)
	go func() {
		defer fnCtx.Close()

		newCtx := NewContext(fnCtx)
		defer newCtx.Exit(nil)

		for i := 0; i < len(nodes) && fnCtx.Accept(); i++ {
			go func() {
				if err := evalContext(newCtx, nodes[i]); err != nil {
					return
				}
			}()

			result, err := newCtx.Output()
			if err != nil {
				return
			}
			fnCtx.Push(result)
		}
	}()

	if !fnCtx.Next() {
		return errors.New("no more stuff")
	}
	value := fnCtx.Argument()

	fnVal, err := ctx.Get(value.raw())
	if err != nil {
		if len(nodes) > 1 {
			return errors.New("unexpected arguments")
		}
		ctx.Yield(value)
		return nil
	}
	fnVal.name = value.raw()

	go func() error {
		_, err = fnVal.Function()(fnCtx)
		if err != nil {
			return err
		}
		return nil
	}()

	collected, err := fnCtx.Collect()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if len(collected) == 1 {
		ctx.Yield(collected[0])
		return nil
	}

	result, err := NewValue(collected)

	ctx.Yield(result)

	return nil
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

		newCtx := NewContext(ctx)
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

		newCtx := NewContext(ctx)
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

		if len(node.Children) < 1 {
			return ctx.Yield(Nil)
		}

		newCtx := NewClosure(ctx)
		go func() {
			defer newCtx.Exit(nil)

			if err := evalContextExpression(newCtx, node.Children); err != nil {
				log.Printf("eval.context: %v", err)
			}
		}()

		for {
			value, err := newCtx.Output()
			if err != nil {
				if err == errClosedChannel {
					return nil
				}
				return err
			}
			ctx.Yield(value)
		}

		return nil
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

	if len(values) == 0 {
		return newCtx, nil, nil
	}

	if len(values) == 1 {
		return newCtx, values[0], nil
	}

	return newCtx, values, nil
}
