package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/sexpr/lexer"
)

func TestNode(t *testing.T) {
	value := &testStringValue{value: "AAAA"}
	token := lexer.NewToken(lexer.TokenBinary, value.Value().(string), 1, 1)

	node := New(token, value)
	_, err := node.PushValue(token, value)
	assert.Error(t, err)
}

func TestNodeList(t *testing.T) {
	value := &testStringValue{value: "["}
	token := lexer.NewToken(lexer.TokenOpenList, value.Value().(string), 1, 1)

	list := NewList(token)
	_, err := list.PushValue(token, value)
	assert.NoError(t, err)
}
