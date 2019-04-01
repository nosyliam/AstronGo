package dc

type ArrayType struct {
	DistributedType

	elemType   BaseType
	arrayRange NumericRange
	arraySize  uint
}

func NewArray(elem BaseType, rng NumericRange) ArrayType {
	a := ArrayType{elemType: elem, arrayRange: rng}
	if rng.IsEmpty() {
		a.arraySize = 0
		a.arrayRange.Min.Uinteger = 0
		a.arrayRange.Max.Uinteger = ^uint64(0)
	} else if rng.Min == rng.Max {
		a.arraySize = uint(rng.Min.Uinteger)
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
	case T_ARRAY, T_BLOB, T_STRING:
		return make([]uint8, a.arraySize)
	case T_VARARRAY, T_VARBLOB, T_VARSTRING:
		return []uint8{}

	}
	return ""
}
func (a ArrayType) GenerateHash(generator HashGenerator) {
	// TODO
}
