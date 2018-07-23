package dc

type ArrayType struct {
	DistributedType

	elemType   BaseType
	arrayRange NumericRange
	arraySize  uint
}

func NewArray(elem BaseType, sz NumericRange) ArrayType {
	a := ArrayType{elemType: elem, arrayRange: sz}
	if sz.IsEmpty() {
		a.arraySize = 0
		a.arrayRange.min.uinteger = 0
		a.arrayRange.max.uinteger = ^uint64(0)
	} else if sz.min == sz.max {
		a.arraySize = uint(sz.min.uinteger)
	} else {
		a.arraySize = 0
	}

	if elem.HasFixedSize() && a.arraySize > 0 {
		a.dtype = T_ARRAY
		a.size = Sizetag_t(uint(elem.GetSize()) * a.arraySize)
	} else {
		a.dtype = T_VARARRAY
		a.size = 0
	}

	if elem.GetType() == T_CHAR {
		if a.dtype == T_ARRAY {
			a.dtype = T_STRING
		} else {
			a.dtype = T_VARSTRING
		}
	} else if elem.GetType() == T_UINT8 {
		if a.dtype == T_ARRAY {
			a.dtype = T_BLOB
		} else {
			a.dtype = T_VARBLOB
		}
	}

	return a
}

func (a ArrayType) GetArraySize() uint       { return a.arraySize }
func (a ArrayType) GetElementType() BaseType { return a.elemType }

func (a ArrayType) HasRange() bool         { return a.arrayRange.IsEmpty() }
func (a ArrayType) GetRange() NumericRange { return a.arrayRange }

func (a ArrayType) GetDefaultValue() interface {} {
	switch a.dtype {
	case T_ARRAY:
	case T_BLOB:
	case T_STRING:
		return [a.arraySize]uint8{}
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
