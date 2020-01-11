package sexpr

import (
	"errors"
	//"log"
)

type symbolTableType uint8

const (
	symbolTableTypeValue symbolTableType = iota
	symbolTableTypeDict
)

type symbolTable struct {
	p *symbolTable

	t symbolTableType
	n map[string]*symbolTable
	v *Value
}

func newSymbolTable(parent *symbolTable) *symbolTable {
	return &symbolTable{
		p: parent,
		t: symbolTableTypeDict,
		n: make(map[string]*symbolTable),
	}
}

func (st *symbolTable) Set(name string, value *Value) error {
	if st.t != symbolTableTypeDict {
		return errors.New("not a dictionary")
	}
	//log.Printf("ST: %p %v -> %v", st, name, value)
	st.n[name] = &symbolTable{
		t: symbolTableTypeValue,
		v: value,
	}
	return nil
}

func (st *symbolTable) Get(name string) (*Value, error) {
	//log.Printf("ST: %p <- %v ??", st, name)
	if st.t != symbolTableTypeDict {
		return nil, errors.New("not a dictionary")
	}
	if value, ok := st.n[name]; ok {
		//log.Printf("ST: %p %v <- %v", st, name, value)
		if value.t == symbolTableTypeValue {
			return value.v, nil
		}
		return nil, errors.New("key is not a value")
	}
	//if st.p != nil {
	//	return st.p.Get(name)
	//}
	return nil, errors.New("no such key")
}
