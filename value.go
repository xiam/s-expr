package sexpr

import (
	"errors"
	"fmt"
)

type ValueType uint8

const (
	ValueTypeNil ValueType = iota
	ValueTypeBinary
	ValueTypeString
	ValueTypeAtom
	ValueTypeInt
	ValueTypeFloat
	ValueTypeBool
	ValueTypeMap
	ValueTypeList
	ValueTypeFunction
)

var valueTypes = map[ValueType]string{
	ValueTypeNil:      "nil",
	ValueTypeBinary:   "binary",
	ValueTypeString:   "string",
	ValueTypeAtom:     "atom",
	ValueTypeInt:      "int",
	ValueTypeFloat:    "float",
	ValueTypeBool:     "bool",
	ValueTypeMap:      "map",
	ValueTypeList:     "list",
	ValueTypeFunction: "function",
}

func (vt ValueType) String() string {
	return valueTypes[vt]
}

type Value struct {
	v interface{}

	Type ValueType
}

var (
	Nil   = &Value{Type: ValueTypeNil}
	True  = &Value{Type: ValueTypeBool, v: true}
	False = &Value{Type: ValueTypeBool, v: false}
)

func NewValue(value interface{}) (*Value, error) {
	switch v := value.(type) {
	case []byte:
		return &Value{v: v, Type: ValueTypeBinary}, nil
	case string:
		return &Value{v: v, Type: ValueTypeString}, nil
	case int64:
		return &Value{v: v, Type: ValueTypeInt}, nil
	case float64:
		return &Value{v: v, Type: ValueTypeFloat}, nil
	case bool:
		return &Value{v: v, Type: ValueTypeBool}, nil
	case map[Value]*Value:
		return &Value{v: v, Type: ValueTypeMap}, nil
	case []*Value:
		return &Value{v: v, Type: ValueTypeList}, nil
	case Function:
		return &Value{v: v, Type: ValueTypeFunction}, nil
	}
	return Nil, errors.New("invalid v")
}

func (v Value) String() string {
	if s, ok := v.v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v.v)
}

func (v Value) Int() int64 {
	return v.v.(int64)
}

func (v Value) Binary() []byte {
	return v.v.([]byte)
}

func (v Value) Float64() float64 {
	return v.v.(float64)
}

func (v Value) Bool() bool {
	return v.v.(bool)
}

func (v Value) Map() map[Value]*Value {
	return v.v.(map[Value]*Value)
}

func (v Value) List() []*Value {
	return v.v.([]*Value)
}

func (v Value) Function() Function {
	return v.v.(Function)
}
