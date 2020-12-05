package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/sexpr/lexer"
)

func TestStringNode(t *testing.T) {
	value := NewStringValue("AAA")
	token := lexer.NewToken(lexer.TokenSequence, value.Value().(string), nil)

	node := NewNode(token, value)
	_, err := node.PushValue(token, value)
	assert.Error(t, err)
}

func TestFloatNode(t *testing.T) {
	value := NewFloatValue(1.234)
	token := lexer.NewToken(lexer.TokenSequence, value.Encode(), nil)

	node := NewNode(token, value)
	_, err := node.PushValue(token, value)
	assert.Error(t, err)
}

func TestNodeList(t *testing.T) {
	value := NewStringValue("[")
	token := lexer.NewToken(lexer.TokenOpenList, value.Value().(string), nil)

	list := NewList(token)
	_, err := list.PushValue(token, value)
	assert.NoError(t, err)
}

func TestNodeMap(t *testing.T) {
	value := NewStringValue("{")
	token := lexer.NewToken(lexer.TokenOpenMap, value.Value().(string), nil)

	list := NewMap(token)
	_, err := list.PushValue(token, value)
	assert.NoError(t, err)
}

func TestNodeExpression(t *testing.T) {
	value := NewStringValue("(")
	token := lexer.NewToken(lexer.TokenOpenExpression, value.Value().(string), nil)

	list := NewExpression(token)
	_, err := list.PushValue(token, value)
	assert.NoError(t, err)
}
