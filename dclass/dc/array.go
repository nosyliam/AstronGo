package dc

type ArrayType struct {
	DistributedType

	elemType   BaseType
	arrayRange NumericRange
	arraySize  uint
}

func NewArray(elem BaseType, rng NumericRange) *ArrayType {
	a := &ArrayType{elemType: elem, arrayRange: rng}
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

func (a *ArrayType) ArraySize() uint       { return a.arraySize }
func (a *ArrayType) ElementType() BaseType { return a.elemType }

func (a *ArrayType) HasRange() bool      { return !a.arrayRange.IsEmpty() }
func (a *ArrayType) Range() NumericRange { return a.arrayRange }
func (a *ArrayType) WithinRange(data []byte, length uint64) bool {
	if a.arraySize > 0 {
		return a.arraySize == uint(length)
	}

	return a.arrayRange.Min.Uinteger < length && a.arrayRange.Max.Uinteger > length
}
func (a *ArrayType) DefaultValue() interface{} {
	switch a.dataType {
	case T_ARRAY, T_BLOB, T_STRING:
		return make([]uint8, a.arraySize)
	case T_VARARRAY, T_VARBLOB, T_VARSTRING:
		return ""

	}
	return ""
}
func (a *ArrayType) GenerateHash(generator *HashGenerator) {
	switch a.dataType {
	case T_ARRAY, T_VARARRAY:
		a.elemType.GenerateHash(generator)
	case T_BLOB, T_VARBLOB:
		if a.elemType.Alias() == "blob" {
			generator.AddInt(int(T_BLOB))
		} else {
			generator.AddInt(int(T_UINT8))
		}

		generator.AddInt(1)
	case T_STRING, T_VARSTRING:
		if a.elemType.Alias() == "string" {
			generator.AddInt(int(T_STRING))
		} else {
			generator.AddInt(int(T_CHAR))
		}

		generator.AddInt(1)
	}

	if a.HasRange() {
		generator.AddInt(1)
		generator.AddInt(int(a.arrayRange.Min.Uinteger))
		generator.AddInt(int(a.arrayRange.Max.Uinteger))
	}
}
