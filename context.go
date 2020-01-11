package sexpr

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

var ctxID = uint64(0)

type Context struct {
	id   uint64
	name string

	Parent *Context

	executable bool

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

func (ctx *Context) Name(name string) *Context {
	ctx.name = name
	log.Printf("*CTX: %v", ctx)
	return ctx
}

func (ctx *Context) IsExecutable() bool {
	return ctx.executable
}

func (ctx *Context) NonExecutable() *Context {
	ctx.executable = false
	return ctx
}

func (ctx *Context) Executable() *Context {
	ctx.executable = true
	return ctx
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
		arg, err := ctx.Argument()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func (ctx *Context) Argument() (*Value, error) {
	if ctx.executable {
		return execArgument(ctx, ctx.lastArgument)
	}
	return ctx.lastArgument, nil
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
	if err := ctx.Yield(values...); err != nil {
		ctx.Exit(err)
		return err
	}

	ctx.Exit(nil)
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

func (ctx *Context) Yield(values ...*Value) error {
	for i := range values {
		if err := ctx.yield(values[i]); err != nil {
			return err
		}
	}
	return nil
}

func (ctx *Context) yield(value *Value) error {
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

func (ctx *Context) String() string {
	return fmt.Sprintf("[%v]: %q (%p)", ctx.id, ctx.name, ctx)
}

func (ctx *Context) Set(name string, value *Value) error {
	if !ctx.executable {
		return errors.New("cannot set on a non-executable context")
	}
	log.Printf("ctx: %v -- %v -> %v", ctx, name, value)
	return ctx.st.Set(name, value)
}

func (ctx *Context) Get(name string) (*Value, error) {
	log.Printf("ctx: %v <- %q ??", ctx, name)
	value, err := ctx.st.Get(name)
	if err != nil {
		if ctx.Parent != nil {
			return ctx.Parent.Get(name)
		}
		return nil, err
	}
	log.Printf("ctx: %v <- %q: %v", ctx, name, value)
	return value, nil
}

func NewClosure(parent *Context) *Context {
	log.Printf("NewClosure")
	ctx := NewContext(parent)
	if parent != nil {
		ctx.st = parent.st
		log.Printf("*map: %v == %v", ctx, parent)
	}
	return ctx
}

func NewContext(parent *Context) *Context {
	log.Printf("NewContext")
	ctx := &Context{
		id:     atomic.AddUint64(&ctxID, 1),
		ticket: make(chan struct{}),

		accept:     make(chan struct{}, 1),
		doneAccept: make(chan struct{}),

		in:  make(chan *Value),
		out: make(chan *Value),
	}
	if parent == nil {
		ctx.st = newSymbolTable(nil)
	} else {
		ctx.Parent = parent
		ctx.executable = parent.executable
		ctx.st = newSymbolTable(parent.st)
	}
	log.Printf("*CTX: %v -> %v", parent, ctx)
	return ctx
}
