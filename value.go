package sexpr

import (
	"fmt"
	"sort"
	"strings"
)

type Function func(*Context) error

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
	v    interface{}
	name string

	Type ValueType
}

var (
	Nil   = &Value{Type: ValueTypeNil}
	True  = &Value{Type: ValueTypeBool, v: true}
	False = &Value{Type: ValueTypeBool, v: false}
)

func NewFunctionValue(v Function) *Value {
	return &Value{v: v, Type: ValueTypeFunction}
}

func NewStringValue(v string) *Value {
	return &Value{v: v, Type: ValueTypeString}
}

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
	return Nil, fmt.Errorf("invalid value %v", value)
}

func (v Value) raw() string {
	return fmt.Sprintf("%v", v.v)
}

func (v Value) String() string {
	switch v.Type {
	case ValueTypeFunction:
		return fmt.Sprintf("<function %v: %#v>", v.name, v.v)
	case ValueTypeString:
		return fmt.Sprintf("%q", v.v.(string))
	case ValueTypeBool:
		t := v.v.(bool)
		if t {
			return ":true"
		}
		return ":false"
	case ValueTypeNil:
		return ":nil"
	case ValueTypeInt:
		return fmt.Sprintf("%d", v.v.(int64))
	case ValueTypeList:
		vv := v.v.([]*Value)
		values := []string{}
		for i := range vv {
			values = append(values, vv[i].String())
		}
		return "[" + strings.Join(values, " ") + "]"
	case ValueTypeMap:
		vv := v.v.(map[Value]*Value)
		values := []string{}
		for k := range vv {
			values = append(values, k.String()+" "+vv[k].String())
		}
		sort.Strings(values)
		return "{" + strings.Join(values, " ") + "}"
	}
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

func compileValue(in interface{}) ([]byte, error) {
	var buf string
	switch v := in.(type) {
	case *Value:
		s := v.String()
		return []byte(s), nil
	case []*Value:
		buf = buf + "["
		for i := range v {
			if i > 0 {
				buf = buf + ", "
			}
			chunk, err := compileValue(v[i])
			if err != nil {
				return nil, err
			}
			buf = buf + string(chunk)
		}
		buf = buf + "]"
	default:
		return nil, fmt.Errorf("unknown type: %T", in)
	}
	return []byte(buf), nil
}
