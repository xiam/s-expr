package ast

// NodeType represents the type of the AST node
type NodeType uint32

// Node types
const (
	nodeTypeValue  NodeType = 1 << 16
	nodeTypeVector          = 1 << 15

	NodeTypeInt    = nodeTypeValue | 1<<0
	NodeTypeFloat  = nodeTypeValue | 1<<1
	NodeTypeSymbol = nodeTypeValue | 1<<2
	NodeTypeAtom   = nodeTypeValue | 1<<3
	NodeTypeString = nodeTypeValue | 1<<4

	NodeTypeList       = nodeTypeVector | 1<<0
	NodeTypeMap        = nodeTypeVector | 1<<1
	NodeTypeExpression = nodeTypeVector | 1<<2
)

func (nt NodeType) String() string {
	s, ok := nodeTypeName[nt]
	if ok {
		return s
	}
	return ""
}

var nodeTypeName = map[NodeType]string{
	NodeTypeInt:        "int",
	NodeTypeFloat:      "float",
	NodeTypeSymbol:     "symbol",
	NodeTypeAtom:       "atom",
	NodeTypeString:     "string",
	NodeTypeList:       "list",
	NodeTypeMap:        "map",
	NodeTypeExpression: "expression",
}
