package dc

type ArrayType struct {
	DistributedType

	elemType   BaseType
	arrayRange NumericRange
	arraySize  uint
}

func NewArray(elem BaseType, rng NumericRange, size uint) ArrayType {
	a := ArrayType{elemType: elem, arrayRange: rng, arraySize: size}
	if rng.IsEmpty() {
		a.arraySize = 0
		a.arrayRange.min.uinteger = 0
		a.arrayRange.max.uinteger = ^uint64(0)
	} else if rng.min == rng.max {
		a.arraySize = uint(rng.min.uinteger)
	} else {
		a.arraySize = 0
	}

	if elem.HasFixedSize() && a.arraySize > 0 {
		a.dataType = T_ARRAY
		a.size = Sizetag_t(uint(elem.Size()) * a.arraySize)
	} else {
		a.dataType = T_VARARRAY
		a.size = 0
	}

	if elem.Type() == T_CHAR {
		if a.dataType == T_ARRAY {
			a.dataType = T_STRING
		} else {
			a.dataType = T_VARSTRING
		}
	} else if elem.Type() == T_UINT8 {
		if a.dataType == T_ARRAY {
			a.dataType = T_BLOB
		} else {
			a.dataType = T_VARBLOB
		}
	}

	return a
}

func (a ArrayType) ArraySize() uint       { return a.arraySize }
func (a ArrayType) ElementType() BaseType { return a.elemType }

func (a ArrayType) HasRange() bool      { return a.arrayRange.IsEmpty() }
func (a ArrayType) Range() NumericRange { return a.arrayRange }

func (a ArrayType) DefaultValue() interface{} {
	switch a.dataType {
	case T_ARRAY:
	case T_BLOB:
	case T_STRING:
		return make([]uint8, a.arraySize)
	case T_VARARRAY:
	case T_VARBLOB:
	case T_VARSTRING:
		return []uint8{}

	}
	return ""
}
func (a ArrayType) GenerateHash(generator HashGenerator) {
	// TODO
}
