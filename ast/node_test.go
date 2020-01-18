package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/sexpr/lexer"
)

func TestStringNode(t *testing.T) {
	value := &testStringValue{value: "AAAA"}
	token := lexer.NewToken(lexer.TokenBinary, value.Value().(string), 1, 1)

	node := New(token, value)
	_, err := node.PushValue(token, value)
	assert.Error(t, err)
}

func TestFloatNode(t *testing.T) {
	value := &testFloatValue{value: 1.234}
	token := lexer.NewToken(lexer.TokenBinary, value.Encode(), 1, 1)

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

func TestNodeMap(t *testing.T) {
	value := &testStringValue{value: "{"}
	token := lexer.NewToken(lexer.TokenOpenMap, value.Value().(string), 1, 1)

	list := NewMap(token)
	_, err := list.PushValue(token, value)
	assert.NoError(t, err)
}

func TestNodeExpression(t *testing.T) {
	value := &testStringValue{value: "("}
	token := lexer.NewToken(lexer.TokenOpenExpression, value.Value().(string), 1, 1)

	list := NewExpression(token)
	_, err := list.PushValue(token, value)
	assert.NoError(t, err)
}
