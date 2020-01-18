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

var (
	_ = Valuer(&testStringValue{})
)
