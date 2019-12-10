package sexpr

import (
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextCreate(t *testing.T) {
	newContext := NewContext(nil)
	assert.NotNil(t, newContext)
}

func TestContextSetGet(t *testing.T) {
	ctx := NewContext(nil)
	assert.NotNil(t, ctx)

	{
		v, err := ctx.Get("foo")
		assert.Error(t, err)
		assert.Nil(t, v)
	}

	{
		err := ctx.Set("foo", True)
		assert.NoError(t, err)

		v, err := ctx.Get("foo")
		assert.NoError(t, err)
		assert.Equal(t, v, True)
	}
}

func TestContextChildContext(t *testing.T) {
	ctx := NewContext(nil)
	assert.NotNil(t, ctx)

	newCtx := NewContext(ctx)

	{
		v, err := ctx.Get("foo")
		assert.Error(t, err)
		assert.Nil(t, v)
	}

	{
		err := ctx.Set("foo", True)
		assert.NoError(t, err)

		v, err := ctx.Get("foo")
		assert.NoError(t, err)
		assert.Equal(t, v, True)
	}

	{
		v, err := newCtx.Get("foo")
		assert.NoError(t, err)
		assert.Equal(t, v, True)
	}
}

func TestContextSequentialInput(t *testing.T) {
	{
		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		go func() {
			defer ctx.Close()
			var err error

			err = ctx.Push(True)
			assert.NoError(t, err)

			err = ctx.Push(False)
			assert.NoError(t, err)

			err = ctx.Push(Nil)
			assert.NoError(t, err)

			err = ctx.Push(False)
			assert.NoError(t, err)
		}()

		{
			for ctx.Next() {
				arg := ctx.Argument()

				log.Printf("arg: %v", *arg)
			}
		}
	}
}

func TestContextInterruptedInput(t *testing.T) {
	{
		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		go func() {
			defer ctx.Close()
			var err error

			err = ctx.Push(True)
			assert.NoError(t, err)

			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			err = ctx.Push(Nil)
			assert.Error(t, err)

			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		{
			for ctx.Next() {
				arg := ctx.Argument()

				log.Printf("arg: %v", *arg)
			}
		}
	}
}

func TestContextEcho(t *testing.T) {
	{
		var wg sync.WaitGroup

		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error

			err = ctx.Push(True)
			assert.NoError(t, err)

			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			err = ctx.Push(Nil)
			assert.Error(t, err)

			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			values, err := ctx.Collect()
			assert.NoError(t, err)

			log.Printf("values: %v", values)
		}()

		{
			for ctx.Next() {
				arg := ctx.Argument()
				assert.NotNil(t, arg)

				err := ctx.Yield(arg)
				assert.NoError(t, err)
			}
			ctx.Exit(nil)
		}

		wg.Wait()
	}
}

func TestContextEcho2(t *testing.T) {
	{
		var wg sync.WaitGroup

		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error

			err = ctx.Push(True)
			assert.NoError(t, err)

			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			err = ctx.Push(Nil)
			assert.Error(t, err)

			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			values, err := ctx.Collect()
			assert.NoError(t, err)

			log.Printf("values: %v", values)
		}()

		{
			args, err := ctx.Arguments()
			assert.NoError(t, err)

			for i := range args {
				err := ctx.Yield(args[i])
				assert.NoError(t, err)
			}
			ctx.Exit(nil)
		}

		wg.Wait()
	}
}

func TestContextEcho3(t *testing.T) {
	{
		var wg sync.WaitGroup

		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error

			err = ctx.Push(True)
			assert.NoError(t, err)

			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			err = ctx.Push(Nil)
			assert.Error(t, err)

			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			values, err := ctx.Collect()
			assert.NoError(t, err)

			log.Printf("values: %v", values)
		}()

		{
			args, err := ctx.Arguments()
			assert.NoError(t, err)

			err = ctx.Return(args...)
			assert.NoError(t, err)
		}

		wg.Wait()
	}
}

func TestContextEcho4(t *testing.T) {
	{
		var wg sync.WaitGroup

		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error

			err = ctx.Push(True)
			assert.NoError(t, err)

			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			err = ctx.Push(Nil)
			assert.Error(t, err)

			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			values, err := ctx.Collect()
			assert.NoError(t, err)

			log.Printf("values: %v", values)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			args, err := ctx.Arguments()
			assert.NoError(t, err)

			err = ctx.Return(args...)
			assert.NoError(t, err)
		}()

		wg.Wait()
	}
}
