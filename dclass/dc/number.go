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
	// TODO: implement psuedo-union
	integer  int64
	uinteger uint64
	float    float64
}

type NumericType struct {
	DistributedType

	divisor     uint32
	origModulus float64
	modulus     Number

	nRange     NumericRange
	nOrigRange NumericRange
}

type NumericRange struct {
	ntype NumberType
	min   Number
	max   Number
}

func NewNumber(ntype Type) NumericType {
	n := NumericType{divisor: 1, nRange: NumericRange{
		min: Number{float: math.Inf(1)},
		max: Number{float: math.Inf(-1)},
	}}
	n.dtype = ntype

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
		n.dtype = T_INVALID
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
		return nil, errors.New("data provided to numeric value is different than datatype size")
	}

	switch n.dtype {
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
		return errors.New("invalid divisor passed to NumericType")
	}

	n.divisor = divisor
	if n.HasRange() {
		n.SetRange(n.nOrigRange)
	}
	if n.HasModulus() {
		n.SetModulus(n.origModulus)
	}

	return nil
}

func (n NumericType) HasModulus() bool    { return n.origModulus != 0 }
func (n NumericType) GetModulus() float64 { return n.origModulus }
func (n NumericType) SetModulus(modulus float64) (err error) {
	uint_mod := uint64(math.Floor(modulus * float64(n.divisor)))
	if modulus <= 0.0 {
		goto invalidModulus
	}

	switch n.dtype {
	case T_CHAR:
	case T_UINT8:
		if uint_mod < 1 || uint64(^uint8(0))+1 < uint_mod {
			goto invalidModulus
		}
		n.modulus.uinteger = uint_mod
	case T_UINT16:
		if uint_mod < 1 || uint64(^uint16(0))+1 < uint_mod {
			goto invalidModulus
		}
		n.modulus.uinteger = uint_mod
	case T_UINT32:
		if uint_mod < 1 || uint64(^uint32(0))+1 < uint_mod {
			goto invalidModulus
		}
		n.modulus.uinteger = uint_mod
	case T_UINT64:
		if uint_mod < 1 {
			goto invalidModulus
		}
		n.modulus.uinteger = uint_mod
	case T_FLOAT32:
	case T_FLOAT64:
		n.modulus.float = modulus * float64(n.divisor)
		break
	default:
		goto invalidModulus
	}

	n.origModulus = modulus
	return nil

invalidModulus:
	return errors.New("invalid modulus passed to NumericType")
}

func (n NumericType) HasRange() bool         { return n.nOrigRange.IsEmpty() }
func (n NumericType) GetRange() NumericRange { return n.nOrigRange }
func (n NumericType) SetRange(rng NumericRange) (err error) {
	n.nOrigRange = rng
	switch n.dtype {
	case T_INT8:
	case T_INT16:
	case T_INT32:
	case T_INT64:
		{
			min := int64(math.Floor(rng.min.float*float64(n.divisor) + 0.5))
			max := int64(math.Floor(rng.max.float*float64(n.divisor) + 0.5))
			n.nRange.min = Number{integer: min}
			n.nRange.max = Number{integer: max}
		}
	case T_CHAR:
	case T_UINT8:
	case T_UINT16:
	case T_UINT32:
	case T_UINT64:
		{
			min := uint64(math.Floor(rng.min.float*float64(n.divisor) + 0.5))
			max := uint64(math.Floor(rng.max.float*float64(n.divisor) + 0.5))
			n.nRange.min = Number{uinteger: min}
			n.nRange.max = Number{uinteger: max}
		}
	case T_FLOAT32:
	case T_FLOAT64:
		{
			n.nRange.min = Number{float: rng.min.float * float64(n.divisor)}
			n.nRange.max = Number{float: rng.max.float * float64(n.divisor)}
		}
	default:
		{
			return errors.New("invalid numeric type for range")
		}
	}

	return nil
}

func (n NumericType) WithinRange(data []byte, length uint64) (ok bool, err error) {
	encoded, err := n.dataToNumber(data)
	if err != nil {
		return false, errors.New("range check failed")
	}
	return n.nRange.Contains(*encoded), nil
}

func (n NumericType) GetDefaultValue() interface{} {
	switch n.dtype {
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
