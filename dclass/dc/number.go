package dc

import (
	"encoding/binary"
	"errors"
	"math"
)

type NumberType int

const (
	NONE NumberType = iota
	INT
	UINT
	FLOAT
)

type Number struct {
	integer  int64
	uinteger uint64
	float    float64
}

type NumericType struct {
	DistributedType

	Divisor uint32
	Modulus float64
	Range   NumericRange

	calculatedRange   NumericRange
	calculatedModulus Number
}

type NumericRange struct {
	ntype NumberType
	min   Number
	max   Number
}

func NewNumber(ntype Type) NumericType {
	n := NumericType{Divisor: 1, calculatedRange: NumericRange{
		min: Number{float: math.Inf(1)},
		max: Number{float: math.Inf(-1)},
	}}
	n.dataType = ntype

	switch ntype {
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
		n.dataType = T_INVALID
	}

	return n
}

func (n NumericRange) Contains(num Number) bool {
	switch n.ntype {
	case NONE:
		return true
	case INT:
		return n.min.integer <= num.integer && num.integer <= n.max.integer
	case UINT:
		return n.min.uinteger <= num.uinteger && num.uinteger <= n.max.uinteger
	case FLOAT:
		return n.min.float <= num.float && num.float <= n.max.float
	default:
		return false
	}
}
func (n NumericRange) IsEmpty() bool { return n.ntype == NONE }

func (n NumericType) dataToNumber(data []byte) (ok *Number, err error) {
	if n.size != Sizetag_t(len(data)) {
		return nil, errors.New("data provided to numeric value is different than type size")
	}

	switch n.dataType {
	case T_INT8:
		return &Number{integer: int64(data[0])}, nil
	case T_INT16:
		return &Number{integer: int64(binary.LittleEndian.Uint16(data))}, nil
	case T_INT32:
		return &Number{integer: int64(binary.LittleEndian.Uint32(data))}, nil
	case T_INT64:
		return &Number{integer: int64(binary.LittleEndian.Uint64(data))}, nil
	case T_CHAR:
	case T_UINT8:
		return &Number{uinteger: uint64(data[0])}, nil
	case T_UINT16:
		return &Number{uinteger: uint64(binary.LittleEndian.Uint16(data))}, nil
	case T_UINT32:
		return &Number{uinteger: uint64(binary.LittleEndian.Uint32(data))}, nil
	case T_UINT64:
		return &Number{uinteger: uint64(binary.LittleEndian.Uint64(data))}, nil
	case T_FLOAT32:
		bits := binary.LittleEndian.Uint32(data)
		return &Number{float: float64(math.Float32frombits(bits))}, nil
	case T_FLOAT64:
		bits := binary.LittleEndian.Uint64(data)
		return &Number{float: math.Float64frombits(bits)}, nil
	}
	return nil, errors.New("unknown error while converting data to number")
}

func (n NumericType) SetDivisor(divisor uint32) (err error) {
	if divisor == 0 {
		return errors.New("invalid Divisor passed to NumericType")
	}

	n.Divisor = divisor
	if !n.Range.IsEmpty() {
		n.SetRange(n.Range)
	}
	if n.Modulus != 0 {
		n.SetModulus(n.Modulus)
	}

	return nil
}

func (n NumericType) SetModulus(modulus float64) (err error) {
	uint_mod := uint64(math.Floor(modulus * float64(n.Divisor)))
	if modulus <= 0.0 {
		goto invalidModulus
	}

	switch n.dataType {
	case T_CHAR:
	case T_UINT8:
		if uint_mod < 1 || uint64(^uint8(0))+1 < uint_mod {
			goto invalidModulus
		}
		n.calculatedModulus.uinteger = uint_mod
		break
	case T_UINT16:
		if uint_mod < 1 || uint64(^uint16(0))+1 < uint_mod {
			goto invalidModulus
		}
		n.calculatedModulus.uinteger = uint_mod
		break
	case T_UINT32:
		if uint_mod < 1 || uint64(^uint32(0))+1 < uint_mod {
			goto invalidModulus
		}
		n.calculatedModulus.uinteger = uint_mod
		break
	case T_UINT64:
		if uint_mod < 1 {
			goto invalidModulus
		}
		n.calculatedModulus.uinteger = uint_mod
		break
	case T_FLOAT32:
	case T_FLOAT64:
		n.calculatedModulus.float = modulus * float64(n.Divisor)
		break
	default:
		goto invalidModulus
	}

	n.Modulus = modulus
	return nil

invalidModulus:
	return errors.New("invalid range for numeric type")
}

func (n NumericType) SetRange(rng NumericRange) (err error) {
	n.Range = rng
	switch n.dataType {
	case T_INT8:
	case T_INT16:
	case T_INT32:
	case T_INT64:
		min := int64(math.Floor(rng.min.float*float64(n.Divisor) + 0.5))
		max := int64(math.Floor(rng.max.float*float64(n.Divisor) + 0.5))
		n.calculatedRange.min = Number{integer: min}
		n.calculatedRange.max = Number{integer: max}
		break
	case T_CHAR:
	case T_UINT8:
	case T_UINT16:
	case T_UINT32:
	case T_UINT64:
		min := uint64(math.Floor(rng.min.float*float64(n.Divisor) + 0.5))
		max := uint64(math.Floor(rng.max.float*float64(n.Divisor) + 0.5))
		n.calculatedRange.min = Number{uinteger: min}
		n.calculatedRange.max = Number{uinteger: max}
		break
	case T_FLOAT32:
	case T_FLOAT64:
		n.calculatedRange.min = Number{float: rng.min.float * float64(n.Divisor)}
		n.calculatedRange.max = Number{float: rng.max.float * float64(n.Divisor)}
		break
	default:
		return errors.New("invalid range for numeric type")
	}

	return nil
}

func (n NumericType) WithinRange(data []byte, length uint64) bool {
	encoded, err := n.dataToNumber(data)
	if err != nil {
		return false
	}
	return n.calculatedRange.Contains(*encoded)
}

func (n NumericType) DefaultValue() interface{} {
	switch n.dataType {
	case T_CHAR:
	case T_INT8:
	case T_UINT8:
		return uint8(0)
	case T_INT16:
	case T_UINT16:
		return uint16(0)
	case T_INT32:
	case T_UINT32:
		return uint32(0)
	case T_FLOAT32:
		return float32(0)
	case T_INT64:
	case T_UINT64:
		return uint64(0)
	case T_FLOAT64:
		return float64(0)
	}

	return 0
}
func (n NumericType) GenerateHash(generator HashGenerator) {
	// TODO
}
