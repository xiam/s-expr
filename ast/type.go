package ast

// NodeType represents the type of the AST node
type NodeType uint16

// Node types
const (
	nodeTypeValue  NodeType = 128
	nodeTypeVector          = 256

	NodeTypeInt    = nodeTypeValue | 1
	NodeTypeFloat  = nodeTypeValue | 2
	NodeTypeSymbol = nodeTypeValue | 4
	NodeTypeAtom   = nodeTypeValue | 8
	NodeTypeString = nodeTypeValue | 16

	NodeTypeList       = nodeTypeVector | 1
	NodeTypeMap        = nodeTypeVector | 2
	NodeTypeExpression = nodeTypeVector | 4
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
