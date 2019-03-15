package parse

import (
	"astrongo/dclass/dc"
	"fmt"
)

func (node ArrayRange) consume() dc.NumericRange {
	return dc.NumericRange{}
}

func (node ArrayBounds) consume() dc.NumericRange {
	return node.ArrayConstraint.consume()
}

func (node Range) consume() dc.NumericRange {
	// TODO
	return dc.NumericRange{}
}

func (node TypeCapture) consume() dc.BaseType {
	var builtType dc.BaseType

	nodeType := dc.StringToType(node.Name)
	switch nodeType {
	case dc.T_INT8:
	case dc.T_INT16:
	case dc.T_INT32:
	case dc.T_INT64:
	case dc.T_UINT8:
	case dc.T_UINT16:
	case dc.T_UINT32:
	case dc.T_UINT64:
	case dc.T_CHAR:
	case dc.T_FLOAT32:
	case dc.T_FLOAT64:
		builtType = dc.NewNumber(nodeType)
	case dc.T_STRING:
		node.Constraint = &Range{}
		builtType = dc.NewArray(dc.NewNumber(dc.T_CHAR), node.Constraint.consume(), 0)
		builtType.SetAlias("string")
	case dc.T_BLOB:
		node.Constraint = &Range{}
		builtType = dc.NewArray(dc.NewNumber(dc.T_UINT8), node.Constraint.consume(), 0)
		builtType.SetAlias("blob")
	}

	return builtType
}

func (node ClassType) traverse(d dc.File) {

}

func (node StructType) traverse(d dc.File) {

}

func (node Typedef) traverse(d dc.File) {

}

func (node TypeDecl) traverse(d dc.File) {
	switch true {
	case node.Keyword != nil:
		d.AddKeyword(node.Keyword.Name)
	case node.Import != nil:
		break
	case node.Typedef != nil:
		break
	case node.Struct != nil:
		break
	case node.Class != nil:
		break
	default:
		panic(fmt.Sprintf("malformed declaration at line %d", node.Pos.Line))
	}
}

func (d DCFile) traverse() {
	file := dc.NewFile()
	for _, declaration := range d.Declarations {
		declaration.traverse(file)
	}
}
