package ast

// Valuer represents a value interface
type Valuer interface {
	Type() NodeType
	Value() interface{}
	Encode() string
}
