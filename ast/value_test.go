package ast

import (
	"fmt"
)

type testStringValue struct {
	value string
}

func (ts *testStringValue) Type() NodeType {
	return NodeTypeString
}

func (ts *testStringValue) Value() interface{} {
	return ts.value
}

func (ts *testStringValue) Encode() string {
	return fmt.Sprintf("%q", ts.value)
}

type testFloatValue struct {
	value float64
}

func (ts *testFloatValue) Type() NodeType {
	return NodeTypeFloat
}

func (ts *testFloatValue) Value() interface{} {
	return ts.value
}

func (ts *testFloatValue) Encode() string {
	return fmt.Sprintf("%0.9f", ts.value)
}

var (
	_ = Valuer(&testStringValue{})
	_ = Valuer(&testFloatValue{})
)
