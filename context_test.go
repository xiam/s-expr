package sexpr

import (
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
			var accept bool

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(True)
			assert.NoError(t, err)

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(False)
			assert.NoError(t, err)

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(Nil)
			assert.NoError(t, err)

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(False)
			assert.NoError(t, err)
		}()

		{
			i := 0
			for ctx.Next() {
				value := ctx.Argument()
				switch i {
				case 0:
					assert.Equal(t, True, value)
				case 1:
					assert.Equal(t, False, value)
				case 2:
					assert.Equal(t, Nil, value)
				case 3:
					assert.Equal(t, False, value)
				}
				i++
			}
			assert.Equal(t, 4, i)
		}
	}
}

func TestContextInterruptedSequentialInput(t *testing.T) {
	{
		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		go func() {
			defer ctx.Close()

			var err error
			var accept bool

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(True)
			assert.NoError(t, err)

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(Nil)
			assert.Error(t, err)

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(False)
			assert.Error(t, err)

			ctx.Close()
		}()

		{
			i := 0
			for ctx.Next() {
				value := ctx.Argument()
				switch i {
				case 0:
					assert.Equal(t, True, value)
				case 1:
					assert.Equal(t, False, value)
				}
				i++
			}
			assert.Equal(t, 2, i)
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
			var accept bool

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(True)
			assert.NoError(t, err)

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(Nil)
			assert.Error(t, err)

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			values, err := ctx.Collect()
			assert.NoError(t, err)

			assert.Equal(t, 2, len(values))
			assert.Equal(t, True, values[0])
			assert.Equal(t, False, values[1])
		}()

		{
			i := 0
			for ctx.Next() {
				value := ctx.Argument()
				assert.NotNil(t, value)

				switch i {
				case 0:
					assert.Equal(t, True, value)
				case 1:
					assert.Equal(t, False, value)
				}

				err := ctx.Yield(value)
				assert.NoError(t, err)

				i++
			}
			assert.Equal(t, 2, i)
			ctx.Exit(nil)
		}

		wg.Wait()
	}
}

func TestContextCollectAllArguments(t *testing.T) {
	{
		var wg sync.WaitGroup

		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		wg.Add(1)
		go func() {
			defer wg.Done()

			var err error
			var accept bool

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(True)
			assert.NoError(t, err)

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(Nil)
			assert.Error(t, err)

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			values, err := ctx.Collect()
			assert.NoError(t, err)

			assert.Equal(t, 2, len(values))
			assert.Equal(t, True, values[0])
			assert.Equal(t, False, values[1])
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

func TestContextReturn(t *testing.T) {
	{
		var wg sync.WaitGroup

		ctx := NewContext(nil)
		assert.NotNil(t, ctx)

		wg.Add(1)
		go func() {
			defer wg.Done()

			var err error
			var accept bool

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(True)
			assert.NoError(t, err)

			accept = ctx.Accept()
			assert.True(t, accept)
			err = ctx.Push(False)
			assert.NoError(t, err)

			ctx.Close()

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(Nil)
			assert.Error(t, err)

			accept = ctx.Accept()
			assert.False(t, accept)
			err = ctx.Push(False)
			assert.Error(t, err)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			values, err := ctx.Collect()
			assert.NoError(t, err)

			assert.Equal(t, 2, len(values))
			assert.Equal(t, True, values[0])
			assert.Equal(t, False, values[1])
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
