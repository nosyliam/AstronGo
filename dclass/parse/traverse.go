package parse

import (
	"astrongo/dclass/dc"
	"fmt"
	"math"
)

func (node Number) consume() int64 {
	val := *node.Value
	if node.Negative {
		val = -val
	}

	return val
}

func (node IntTransform) apply(n *dc.NumericType) {
	switch node.Operator {
	case "%":
		if ok := n.SetModulus(float64(node.Value)); !ok {
			goto fail
		}
	case "/":
		if ok := n.SetDivisor(uint32(node.Value)); !ok {
			goto fail
		}
	default:
		goto fail
	}

	return

fail:
	panic(fmt.Sprintf("invalid integer transformation at line %d, column %d", node.Pos.Line, node.Pos.Column))
}

func (node ArrayBounds) consume(n dc.BaseType) dc.BaseType {
	if node.ArrayConstraint == nil {
		node.ArrayConstraint = &ArrayRange{}
	}

	return dc.NewArray(n, node.ArrayConstraint.consume())
}

func (node ArrayRange) consume() dc.NumericRange {
	var lo, hi float64

	if node.Lo == nil && node.Hi == nil {
		return dc.NumericRange{Type: dc.NONE}
	}

	if node.Lo != nil {
		lo = float64(*node.Lo)
	} else {
		lo = math.Inf(-1)
	}

	if node.Hi != nil {
		hi = float64(*node.Hi)
	} else if node.Lo != nil {
		hi = lo
	} else {
		hi = math.Inf(1)
	}

	return dc.NumericRange{Type: dc.INT,
		Min: dc.Number{Integer: int64(lo), Uinteger: uint64(lo), Float: lo},
		Max: dc.Number{Integer: int64(hi), Uinteger: uint64(hi), Float: hi}}
}

func (node Range) consume(dtype dc.Type) dc.NumericRange {
	var ntype dc.NumberType
	var lo, hi float64

	if node.Lo == nil && node.Hi == nil {
		return dc.NumericRange{Type: dc.NONE}
	}

	switch dtype {
	case dc.T_INT8, dc.T_INT16, dc.T_INT32, dc.T_INT64, dc.T_CHAR:
		ntype = dc.INT
	case dc.T_UINT8, dc.T_UINT16, dc.T_UINT32, dc.T_UINT64, dc.T_STRING, dc.T_BLOB:
		ntype = dc.UINT
	case dc.T_FLOAT32, dc.T_FLOAT64:
		ntype = dc.FLOAT
	}

	if node.Lo != nil {
		lo = float64(node.Lo.consume())
	} else {
		lo = math.Inf(-1)
	}

	if node.Hi != nil {
		hi = float64(node.Hi.consume())
	} else if node.Lo != nil {
		hi = lo
	} else {
		hi = math.Inf(1)
	}

	return dc.NumericRange{Type: ntype,
		Min: dc.Number{Integer: int64(lo), Uinteger: uint64(lo), Float: lo},
		Max: dc.Number{Integer: int64(hi), Uinteger: uint64(hi), Float: hi}}
}

func (node ArrayValue) consume(dataType dc.Type) []interface{} {
	var value []interface{}
	if node.String != nil {
		value = append(value, node.String)
	} else {
		var convNum interface{}
		number := node.Number.consume()
		switch dataType {
		case dc.T_CHAR, dc.T_INT8, dc.T_UINT8:
			convNum = uint8(number)
		case dc.T_INT16, dc.T_UINT16:
			convNum = uint16(number)
		case dc.T_INT32, dc.T_UINT32:
			convNum = uint32(number)
		case dc.T_FLOAT32:
			convNum = float32(number)
		case dc.T_INT64, dc.T_UINT64:
			convNum = uint64(number)
		case dc.T_FLOAT64:
			convNum = float64(number)
		default:
			convNum = uint8(number)
		}

		if node.Multiplier != nil {
			for i := 0; i < int(*node.Multiplier); i++ {
				value = append(value, convNum)
			}
		} else {
			value = append(value, convNum)
		}
	}

	return value
}

func (node DefaultValue) consume(dataType dc.Type) []interface{} {
	var defaultValue []interface{}

	switch {
	case node.Array:
		for _, array := range node.ArrayDefault {
			defaultValue = append(defaultValue, array.consume(dataType))
		}
	case node.Integer != nil:
		val := *node.Integer
		if node.Negative {
			val = -val
		}

		defaultValue = append(defaultValue, val)
	case node.Float != nil:
		val := *node.Float
		if node.Negative {
			val = -val
		}

		defaultValue = append(defaultValue, val)
	case node.String != nil:
		defaultValue = append(defaultValue, *node.String)
	}

	return defaultValue
}

func (node TypeCapture) consume(d *dc.File) dc.BaseType {
	var builtType dc.BaseType

	if node.Constraint == nil {
		node.Constraint = &Range{}
	}

	nodeType := dc.StringToType(node.Name)
	switch nodeType {
	case dc.T_INT8, dc.T_INT16, dc.T_INT32, dc.T_INT64, dc.T_UINT8, dc.T_UINT16, dc.T_UINT32,
		dc.T_UINT64, dc.T_CHAR, dc.T_FLOAT32, dc.T_FLOAT64:
		num := dc.NewNumber(nodeType)
		for _, trans := range node.Transforms {
			trans.apply(num)
		}

		num.SetRange(node.Constraint.consume(nodeType))
		builtType = num
	case dc.T_STRING:
		builtType = dc.NewArray(dc.NewNumber(dc.T_CHAR), node.Constraint.consume(dc.T_CHAR))
		builtType.SetAlias("string")
	case dc.T_BLOB:
		builtType = dc.NewArray(dc.NewNumber(dc.T_UINT8), node.Constraint.consume(dc.T_UINT8))
		builtType.SetAlias("blob")
	default:
		if ntype, ok := d.TypeByName(node.Name); ok {
			builtType = *ntype
			if _, ok := builtType.(*dc.Method); ok {
				panic(fmt.Sprintf("ambiguous type cannot be a method at line %d", node.Pos.Line))
			}
		} else {
			panic(fmt.Sprintf("type '%s' has not been declared at line %d", node.Name, node.Pos.Line))
		}
	}

	for _, bounds := range node.Bounds {
		builtType = bounds.consume(builtType)
	}

	return builtType
}

func (node Typedef) traverse(d *dc.File) {
	var base dc.BaseType
	base = node.Base.consume(d)

	for _, bounds := range node.Type.Bounds {
		base = bounds.consume(base)
	}

	base.SetAlias(node.Type.Name)
	if err := d.AddTypedef(node.Type.Name, base); err != nil {
		panic(fmt.Sprintf("cannot add typedef '%s' at line %d as a type was already declared with that name", node.Type.Name, node.Pos.Line))
	}
}

func (node IntParameter) consume() dc.BaseType {
	var builtType dc.BaseType
	nodeType := dc.StringToType(node.Type)
	numType := dc.NewNumber(nodeType)

	if node.Constraint == nil {
		node.Constraint = &Range{}
	}

	for _, trans := range node.Transforms {
		trans.apply(numType)
	}

	numType.SetRange(node.Constraint.consume(nodeType))
	if node.ArrayPrefix != nil && node.ArraySuffix != nil {
		panic(fmt.Sprintf("invalid syntax at line %d", node.Pos.Line))
	}

	builtType = dc.BaseType(numType)
	for _, bounds := range node.ArraySuffix {
		builtType = bounds.consume(builtType)
	}

	for _, bounds := range node.ArrayPrefix {
		builtType = bounds.consume(builtType)
	}

	return dc.BaseType(builtType)
}

func (node FloatParameter) consume() dc.BaseType {
	var builtType dc.BaseType
	nodeType := dc.StringToType(node.Type)
	numType := dc.NewNumber(nodeType)

	if node.Constraint == nil {
		node.Constraint = &Range{}
	}

	for _, trans := range node.Transforms {
		trans.apply(numType)
	}

	numType.SetRange(node.Constraint.consume(nodeType))
	if node.ArrayPrefix != nil && node.ArraySuffix != nil {
		panic(fmt.Sprintf("invalid syntax at line %d", node.Pos.Line))
	}

	builtType = dc.BaseType(numType)
	for _, bounds := range node.ArraySuffix {
		builtType = bounds.consume(builtType)
	}

	for _, bounds := range node.ArrayPrefix {
		builtType = bounds.consume(builtType)
	}

	return dc.BaseType(builtType)
}

func (node CharParameter) consume() dc.BaseType {
	var builtType dc.BaseType
	charType := dc.NewNumber(dc.T_CHAR)

	if node.ArrayPrefix != nil && node.ArraySuffix != nil {
		panic(fmt.Sprintf("invalid syntax at line %d", node.Pos.Line))
	}

	builtType = dc.BaseType(charType)
	for _, bounds := range node.ArraySuffix {
		builtType = bounds.consume(builtType)
	}

	for _, bounds := range node.ArrayPrefix {
		builtType = bounds.consume(builtType)
	}

	return dc.BaseType(builtType)
}

func (node SizedParameter) consume() dc.BaseType {
	var builtType dc.BaseType
	nodeType := dc.StringToType(node.Type)
	var elemType dc.BaseType
	switch nodeType {
	case dc.T_STRING:
		elemType = dc.NewNumber(dc.T_CHAR)
		elemType.SetAlias("string")
	case dc.T_BLOB:
		elemType = dc.NewNumber(dc.T_UINT8)
		elemType.SetAlias("blob")
	}

	if node.Constraint == nil {
		node.Constraint = &Range{}
	}

	sizedType := dc.NewArray(elemType, node.Constraint.consume(nodeType))

	if node.ArrayPrefix != nil && node.ArraySuffix != nil {
		panic(fmt.Sprintf("invalid syntax at line %d", node.Pos.Line))
	}

	builtType = dc.BaseType(sizedType)
	for _, bounds := range node.ArraySuffix {
		builtType = bounds.consume(builtType)
	}

	for _, bounds := range node.ArrayPrefix {
		builtType = bounds.consume(builtType)
	}

	return dc.BaseType(builtType)
}

func (node AmbiguousParameter) consume(d *dc.File) dc.BaseType {
	var builtType dc.BaseType

	if ntype, ok := d.TypeByName(node.Type); ok {
		builtType = *ntype
	} else {
		panic(fmt.Sprintf("type '%s' has not been declared at line %d", node.Type, node.Pos.Line))
	}

	if _, ok := builtType.(*dc.Method); ok {
		panic(fmt.Sprintf("ambiguous type cannot be a method at line %d", node.Pos.Line))
	}

	builtType = dc.BaseType(builtType)
	for _, bounds := range node.ArrayConstraint {
		builtType = bounds.consume(builtType)
	}

	return dc.BaseType(builtType)
}

func (node Parameter) consume(d *dc.File) dc.BaseType {
	var builtType dc.BaseType

	switch {
	case node.Float != nil:
		builtType = node.Float.consume()
	case node.Char != nil:
		builtType = node.Char.consume()
	case node.Int != nil:
		builtType = node.Int.consume()
	case node.Sized != nil:
		builtType = node.Sized.consume()
	case node.Typed != nil:
		builtType = node.Typed.consume(d)
	}

	return builtType
}

func (node Parameter) name() string {
	var str *string

	switch {
	case node.Float != nil:
		str = node.Float.Identifier
	case node.Char != nil:
		str = node.Char.Identifier
	case node.Int != nil:
		str = node.Int.Identifier
	case node.Sized != nil:
		str = node.Sized.Identifier
	case node.Typed != nil:
		str = node.Typed.Identifier
	}

	if str == nil {
		return ""
	} else {
		return *str
	}
}

func (node Parameter) defaultValue(d *dc.File) []interface{} {
	var defaultValue []interface{}

	switch {
	case node.Float != nil && node.Float.Default != nil:
		defaultValue = node.Float.Default.consume(dc.StringToType(node.Float.Type))
	case node.Char != nil && node.Char.Default != nil:
		defaultValue = node.Char.Default.consume(dc.StringToType(node.Char.Type))
	case node.Int != nil && node.Int.Default != nil:
		defaultValue = node.Int.Default.consume(dc.StringToType(node.Int.Type))
	case node.Sized != nil && node.Sized.Default != nil:
		defaultValue = node.Sized.Default.consume(dc.StringToType(node.Sized.Type))
	case node.Typed != nil && node.Typed.Default != nil:
		defaultValue = node.Typed.Default.consume(node.Typed.consume(d).Type())
	}

	return defaultValue
}

func (node MethodField) consume(d *dc.File) dc.Field {
	var defaultValue []interface{}
	method := dc.NewMethod()
	for _, p := range node.Parameters {
		param := dc.NewParameter(method)
		if err := param.SetName(p.name()); err != nil {
			panic(fmt.Sprintf("%s at line %d column %d", err.Error(), p.Pos.Line, p.Pos.Column))
		}

		if err := param.SetType(p.consume(d)); err != nil {
			panic(fmt.Sprintf("%s at line %d column %d", err.Error(), p.Pos.Line, p.Pos.Column))
		}

		if err := method.AddParameter(param); err != nil {
			panic(fmt.Sprintf("%s at line %d", err.Error(), p.Pos.Line))
		}

		defaultValue = append(defaultValue, p.defaultValue(d))
	}

	field := dc.NewAtomicField(method, node.Name)
	field.SetDefaultValue(defaultValue)

	if node.Keywords != nil {
		for _, keyword := range node.Keywords.Keywords {
			field.(*dc.AtomicField).AddKeyword(keyword)
		}
	}

	return field
}

func (node MolecularField) consume(cls *dc.Class) dc.Field {
	field := dc.NewMolecularField(node.Name)
	for _, child := range node.Fields {
		if f, ok := cls.GetFieldByName(child); ok {
			if err := field.AddField(f); err != nil {
				panic(fmt.Sprintf("%s at line %d", err.Error(), node.Pos.Line))
			}
		} else {
			panic(fmt.Sprintf("unknown molecular field %s at line %d", child, node.Pos.Line))
		}
	}

	return field
}

func (node AtomicField) consume(d *dc.File) dc.Field {
	field := dc.NewAtomicField(node.Parameter.consume(d), node.Parameter.name())
	field.SetDefaultValue(node.Parameter.defaultValue(d))

	if node.Keywords != nil {
		for _, keyword := range node.Keywords.Keywords {
			field.(*dc.AtomicField).AddKeyword(keyword)
		}
	}

	return field
}

func (node FieldDecl) consume(d *dc.File, cls *dc.Class) dc.Field {
	var builtType dc.Field

	switch true {
	case node.Method != nil:
		builtType = node.Method.consume(d)
	case node.Molecular != nil:
		builtType = node.Molecular.consume(cls)
	case node.Atomic != nil:
		builtType = node.Atomic.consume(d)
	}

	return builtType
}

func (node ClassType) traverse(d *dc.File) {
	class := dc.NewClass(node.Name, d)
	for _, parent := range node.Parents {
		if parentClass, ok := d.ClassByName(parent); ok {
			class.AddParent(*parentClass)
		} else {
			panic(fmt.Sprintf("parent class '%s' has not been declared at line %d", parent, node.Pos.Line))
		}
	}

	for _, field := range node.Declarations {
		if err := class.AddField(field.consume(d, class)); err != nil {
			panic(fmt.Sprintf("%s at line %d", err.Error(), field.Pos.Line))
		}
	}

	if err := d.AddClass(class); err != nil {
		panic(fmt.Sprintf("cannot add class '%s' at line %d as a type was already declared with that name", node.Name, node.Pos.Line))
	}
}

func (node StructType) traverse(d *dc.File) {
	strct := dc.NewStruct(node.Name, d)
	for _, field := range node.Declarations {
		if err := strct.AddField(field.consume(d, nil)); err != nil {
			panic(fmt.Sprintf("%s at line %d", err.Error(), field.Pos.Line))
		}
	}

	if err := d.AddStruct(strct); err != nil {
		panic(fmt.Sprintf("cannot add struct '%s' at line %d as a type was already declared with that name", node.Name, node.Pos.Line))
	}
}

func (node TypeDecl) traverse(d *dc.File) {
	switch {
	case node.Keyword != nil:
		d.AddKeyword(node.Keyword.Name)
	case node.Import != nil:
	case node.Typedef != nil:
		node.Typedef.traverse(d)
	case node.Struct != nil:
		node.Struct.traverse(d)
	case node.Class != nil:
		node.Class.traverse(d)
	default:
		panic(fmt.Sprintf("malformed declaration at line %d", node.Pos.Line))
	}
}

func (d DCFile) Traverse() *dc.File {
	file := dc.NewFile()

	for _, declaration := range d.Declarations {
		declaration.traverse(file)
	}

	return file
}
