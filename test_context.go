package sexpr

import (
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
		assert.Error(t, err)

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
		assert.Error(t, err)

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

func TestContextInput(t *testing.T) {
	ctx := NewContext(nil)
	assert.NotNil(t, ctx)

	{
		for ctx.Next() {
			ctx.NextInput()
		}

		err := ctx.Get("foo")
		assert.Error(t, err)
		assert.Nil(t, v)
	}

}
