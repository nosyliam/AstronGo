package parse

import (
	"astrongo/dclass/dc"
	"fmt"
	"math"
)

func (node IntTransform) apply(n dc.NumericType) {
	switch node.Operator {
	case "%":
		if ok := n.SetModulus(float64(node.Value)); !ok {
			goto fail
		}
		break
	case "/":
		if ok := n.SetDivisor(uint32(node.Value)); !ok {
			goto fail
		}
	default:
		goto fail
	}

fail:
	panic(fmt.Sprintf("invalid integer transformation at line %d, column", node.Pos.Line, node.Pos.Column))
}

func (node ArrayRange) consume() dc.NumericRange {
	return dc.NumericRange{}
}

func (node ArrayBounds) consume() dc.NumericRange {
	return node.ArrayConstraint.consume()
}

func (node Range) consume(dtype dc.Type) dc.NumericRange {
	var ntype dc.NumberType
	var lo, hi float64

	switch dtype {
	case dc.T_INT8:
	case dc.T_INT16:
	case dc.T_INT32:
	case dc.T_INT64:
	case dc.T_CHAR:
		ntype = dc.INT
		break
	case dc.T_UINT8:
	case dc.T_UINT16:
	case dc.T_UINT32:
	case dc.T_UINT64:
		ntype = dc.UINT
	case dc.T_FLOAT32:
	case dc.T_FLOAT64:
		ntype = dc.FLOAT
		break
	}

	if node.Lo != nil {
		lo = float64(*node.Lo.Value)
	} else {
		lo = math.Inf(1)
	}

	if node.Hi != nil {
		lo = float64(*node.Hi.Value)
	} else {
		lo = math.Inf(-1)
	}

	return dc.NumericRange{Type: ntype,
		Min: dc.Number{Integer: int64(lo), Uinteger: uint64(lo), Float: lo},
		Max: dc.Number{Integer: int64(hi), Uinteger: uint64(hi), Float: hi}}
}

func (node TypeCapture) consume(d dc.File) dc.BaseType {
	var builtType dc.BaseType

	if node.Constraint == nil {
		node.Constraint = &Range{}
	}

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
		num := dc.NewNumber(nodeType)
		for _, trans := range node.Transforms {
			trans.apply(num)
		}

		num.SetRange(node.Constraint.consume(nodeType))
		builtType = num
		break
	case dc.T_STRING:
		builtType = dc.NewArray(dc.NewNumber(dc.T_CHAR), node.Constraint.consume(dc.T_CHAR), 0)
		builtType.SetAlias("string")
		break
	case dc.T_BLOB:
		builtType = dc.NewArray(dc.NewNumber(dc.T_UINT8), node.Constraint.consume(dc.T_UINT8), 0)
		builtType.SetAlias("blob")
	default:
		if ntype, ok := d.TypeByName(node.Name); ok {
			builtType = *ntype
		} else {
			panic(fmt.Sprintf("type '%s' has not been declared at line %d", node.Name, node.Pos.Line))
		}
	}

	return builtType
}

func (node Typedef) traverse(d dc.File) {
	base := node.Base.consume(d)
	if _, ok := base.(dc.ArrayType); ok {
		panic(fmt.Sprintf("typedef '%s' references an array type at line %d", node.Base.Name, node.Pos.Line))
	}

}

func (node ClassType) traverse(d dc.File) {

}

func (node StructType) traverse(d dc.File) {

}

func (node TypeDecl) traverse(d dc.File) {
	switch true {
	case node.Keyword != nil:
		d.AddKeyword(node.Keyword.Name)
	case node.Import != nil:
		break
	case node.Typedef != nil:
		node.Typedef.traverse(d)
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
