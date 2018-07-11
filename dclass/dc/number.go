package dc

import (
	"errors"
	"encoding/binary"
	"math"
	"fmt"
)

type NumberType int
const (
	NONE NumberType = iota
	INT
	UINT
	FLOAT
)

type Number struct {
	// TODO: implement psuedo-union
	integer int64
	uinteger uint64
	float float64
}

type NumericType struct {
	DistributedType

	divisor uint32
	origModulus float64
	modulus Number

	nRange NumericRange
	nOrigRange NumericRange
}

type NumericRange struct {
	ntype NumberType
	min Number
	max Number
}

func NewNumber(ntype Type) NumericType {
	n := NumericType{divisor: 1, nRange: NumericRange{
		min: Number{float: math.Inf(1)},
		max: Number{float: math.Inf(-1)},
	}}
	n.dtype = ntype

	switch (ntype) {
	case T_CHAR:
	case T_INT8:
	case T_UINT8:
		n.size = 1
	case T_INT16:
	case T_UINT16:
		n.size = 2
	case T_INT32:
	case T_UINT32:
	case T_FLOAT32:
		n.size = 4
	case T_INT64:
	case T_UINT64:
	case T_FLOAT64:
		n.size = 8
	default:
		n.dtype = T_INVALID
	}

	return n
}

func (n *NumericRange) Contains(num Number) bool {
	switch (n.ntype) {
	case NONE:
		return true
	case INT:
		return n.min.integer <= num.integer && num.integer <= n.max.integer
	case UINT:
		return n.min.uinteger <= num.uinteger && num.uinteger <= n.max.uinteger
	case FLOAT:
		return n.min.float <= num.float && num.float <= n.max.float;
	default:
		return false
	}
}
func (n *NumericRange) IsRangeEmpty() bool { return n.ntype == NONE }

func (n *NumericType) DataToNumber(data []byte) (err error, ok Number) {
	if (n.size != Sizetag_t(len(data))) {
		return errors.New("data provided to numeric value is different than datatype size"), Number{}
	}

	switch(n.dtype) {
	case T_INT8:
		return nil, Number{integer: int64(data[0])}
	case T_INT16:
		return nil, Number{integer: int64(binary.LittleEndian.Uint16(data))}
	case T_INT32:
		return nil, Number{integer: int64(binary.LittleEndian.Uint32(data))}
	case T_INT64:
		return nil, Number{integer: int64(binary.LittleEndian.Uint64(data))}
	case T_CHAR:
	case T_UINT8:
		return nil, Number{uinteger: uint64(data[0])}
	case T_UINT16:
		return nil, Number{uinteger: uint64(binary.LittleEndian.Uint16(data))}
	case T_UINT32:
		return nil, Number{uinteger: uint64(binary.LittleEndian.Uint32(data))}
	case T_UINT64:
		return nil, Number{uinteger: uint64(binary.LittleEndian.Uint64(data))}
	case T_FLOAT32:
		bits := binary.LittleEndian.Uint32(data)
		return nil, Number{float: float64(math.Float32frombits(bits))}
	case T_FLOAT64:
		bits := binary.LittleEndian.Uint64(data)
		return nil, Number{float: math.Float64frombits(bits)}
	}
	return errors.New("unknown error while converting data to number"), Number{}
}


func (n *NumericType) SetDivisor(divisor uint32) bool {
	if divisor == 0 {
		return false
	}

	n.divisor = divisor
	if n.HasRange() {
		n.SetRange(n.nOrigRange)
	}
	if n.HasModulus() {
		n.SetModulus(n.origModulus)
	}
}

func (n *NumericType) HasModulus() bool { return n.origModulus != 0 }
func (n *NumericType) GetModulus() float64 { return n.origModulus }
func (n *NumericType) SetModulus(modulus float64) bool {
	if modulus <= 0.0 {
		return false
	}

	uint_mod := uint64(math.Floor(modulus * float64(n.divisor)))

	switch (n.dtype) {
	case T_CHAR:
	case T_UINT8:
		if uint_mod < 1 || uint64(^uint8(0)) + 1 < uint_mod {
			return false;
		}
		n.modulus.uinteger = uint_mod
	case T_UINT16:
		if uint_mod < 1 || uint64(^uint16(0)) + 1 < uint_mod {
			return false;
		}
		n.modulus.uinteger = uint_mod;
	case T_UINT32:
		if uint_mod < 1 || uint64(^uint32(0)) + 1 < uint_mod {
			return false;
	}
		n.modulus.uinteger = uint_mod;
	case T_UINT64:
		if(uint_mod < 1) {
			return false;
		}
		n.modulus.uinteger = uint_mod;
	case T_FLOAT32:
	case T_FLOAT64:
		n.modulus.float = modulus * float64(n.divisor)
		break;
	default:
		return false;
	}

	n.origModulus = modulus
	return true
}

func (n *NumericType) SetRange(rng NumericRange) bool {
	if rng.ntype != FLOAT {
		return false;
	}

	n.nOrigRange = rng;
	switch(n.dtype) {
	case T_INT8:
	case T_INT16:
	case T_INT32:
	case T_INT64: {
		min := int64(math.Floor(rng.min.float * float64(n.divisor) + 0.5))
		max := int64(math.Floor(rng.max.float * float64(n.divisor) + 0.5));
		n.nRange.min = Number{integer: min}
		n.nRange.max = Number{integer: max}
	}
	case T_CHAR:
	case T_UINT8:
	case T_UINT16:
	case T_UINT32:
	case T_UINT64: {
		min := uint64(math.Floor(rng.min.float * float64(n.divisor) + 0.5))
		max := uint64(math.Floor(rng.max.float * float64(n.divisor) + 0.5));
		n.nRange.min = Number{uinteger: min}
		n.nRange.max = Number{uinteger: max}
	}
	case T_FLOAT32:
	case T_FLOAT64: {
		n.nRange.min = Number{float: rng.min.float * float64(n.divisor)}
		n.nRange.max = Number{float: rng.max.float * float64(n.divisor)}
	}
	default: {
		return false;
	}
	}

	return true;
}

func (n *NumericType) WithinRange(data []byte, length uint64) bool {
	if ok, encoded := n.DataToNumber(data); ok == nil {
		return n.nRange.Contains(encoded)
	}
	// TODO: add logging for data unpacking error
	return false
}