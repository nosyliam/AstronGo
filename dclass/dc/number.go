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
	Integer  int64
	Uinteger uint64
	Float    float64
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
	Type NumberType
	Min  Number
	Max  Number
}

func NewNumber(ntype Type) *NumericType {
	n := &NumericType{Divisor: 1, calculatedRange: NumericRange{
		Min: Number{Float: math.Inf(-1)},
		Max: Number{Float: math.Inf(1)},
	}}
	n.dataType = ntype

	switch ntype {
	case T_CHAR, T_INT8, T_UINT8:
		n.size = 1
	case T_INT16, T_UINT16:
		n.size = 2
	case T_INT32, T_UINT32, T_FLOAT32:
		n.size = 4
	case T_INT64, T_UINT64, T_FLOAT64:
		n.size = 8
	default:
		n.dataType = T_INVALID
	}

	return n
}

func (n NumericRange) Contains(num Number) bool {
	switch n.Type {
	case NONE:
		return true
	case INT:
		return n.Min.Integer <= num.Integer && num.Integer <= n.Max.Integer
	case UINT:
		return n.Min.Uinteger <= num.Uinteger && num.Uinteger <= n.Max.Uinteger
	case FLOAT:
		return n.Min.Float <= num.Float && num.Float <= n.Max.Float
	default:
		return false
	}
}
func (n NumericRange) IsEmpty() bool { return n.Type == NONE }

func (n NumericType) dataToNumber(data []byte) (ok *Number, err error) {
	if n.size != Sizetag_t(len(data)) {
		return nil, errors.New("data provided to numeric value is different than type size")
	}

	switch n.dataType {
	case T_INT8:
		return &Number{Integer: int64(data[0])}, nil
	case T_INT16:
		return &Number{Integer: int64(binary.LittleEndian.Uint16(data))}, nil
	case T_INT32:
		return &Number{Integer: int64(binary.LittleEndian.Uint32(data))}, nil
	case T_INT64:
		return &Number{Integer: int64(binary.LittleEndian.Uint64(data))}, nil
	case T_CHAR, T_UINT8:
		return &Number{Uinteger: uint64(data[0])}, nil
	case T_UINT16:
		return &Number{Uinteger: uint64(binary.LittleEndian.Uint16(data))}, nil
	case T_UINT32:
		return &Number{Uinteger: uint64(binary.LittleEndian.Uint32(data))}, nil
	case T_UINT64:
		return &Number{Uinteger: uint64(binary.LittleEndian.Uint64(data))}, nil
	case T_FLOAT32:
		bits := binary.LittleEndian.Uint32(data)
		return &Number{Float: float64(math.Float32frombits(bits))}, nil
	case T_FLOAT64:
		bits := binary.LittleEndian.Uint64(data)
		return &Number{Float: math.Float64frombits(bits)}, nil
	}
	return nil, errors.New("unknown error while converting data to number")
}

func (n *NumericType) SetDivisor(divisor uint32) (ok bool) {
	if divisor == 0 {
		return false
	}

	n.Divisor = divisor
	if !n.Range.IsEmpty() {
		n.SetRange(n.Range)
	}
	if n.Modulus != 0 {
		n.SetModulus(n.Modulus)
	}

	return true
}

func (n *NumericType) SetModulus(modulus float64) (ok bool) {
	uint_mod := uint64(math.Floor(modulus*float64(n.Divisor) + 0.5))
	if modulus <= 0 {
		return false
	}

	switch n.dataType {
	case T_CHAR, T_UINT8, T_INT8:
		if uint_mod < 1 || uint64(^uint8(0))+1 < uint_mod {
			return false
		}
		n.calculatedModulus.Uinteger = uint_mod
		break
	case T_UINT16, T_INT16:
		if uint_mod < 1 || uint64(^uint16(0))+1 < uint_mod {
			return false
		}
		n.calculatedModulus.Uinteger = uint_mod
		break
	case T_UINT32, T_INT32:
		if uint_mod < 1 || uint64(^uint32(0))+1 < uint_mod {
			return false
		}
		n.calculatedModulus.Uinteger = uint_mod
		break
	case T_UINT64, T_INT64:
		if uint_mod < 1 {
			return false
		}
		n.calculatedModulus.Uinteger = uint_mod
		break
	case T_FLOAT32, T_FLOAT64:
		n.calculatedModulus.Float = modulus * float64(n.Divisor)
		break
	default:
		return false
	}

	n.Modulus = modulus
	return true
}

func (n *NumericType) SetRange(rng NumericRange) {
	n.Range = rng
	switch n.dataType {
	case T_INT8, T_INT16, T_INT32, T_INT64:
		min := int64(math.Floor(rng.Min.Float*float64(n.Divisor) + 0.5))
		max := int64(math.Floor(rng.Max.Float*float64(n.Divisor) + 0.5))
		n.calculatedRange.Min = Number{Integer: min}
		n.calculatedRange.Max = Number{Integer: max}
		n.calculatedRange.Type = INT
		break
	case T_CHAR, T_UINT8, T_UINT16, T_UINT32, T_UINT64:
		min := uint64(math.Floor(rng.Min.Float*float64(n.Divisor) + 0.5))
		max := uint64(math.Floor(rng.Max.Float*float64(n.Divisor) + 0.5))
		n.calculatedRange.Min = Number{Uinteger: min}
		n.calculatedRange.Max = Number{Uinteger: max}
		n.calculatedRange.Type = UINT
		break
	case T_FLOAT32, T_FLOAT64:
		n.calculatedRange.Min = Number{Float: rng.Min.Float * float64(n.Divisor)}
		n.calculatedRange.Max = Number{Float: rng.Max.Float * float64(n.Divisor)}
		n.calculatedRange.Type = FLOAT
		break
	}
}

func (n NumericType) WithinRange(data []byte, length uint64) bool {
	encoded, err := n.dataToNumber(data)
	if err != nil {
		return false
	}
	return n.calculatedRange.Contains(*encoded)
}

func (n *NumericType) DefaultValue() interface{} {
	switch n.dataType {
	case T_CHAR, T_INT8, T_UINT8:
		return uint8(0)
	case T_INT16, T_UINT16:
		return uint16(0)
	case T_INT32, T_UINT32:
		return uint32(0)
	case T_FLOAT32:
		return float32(0)
	case T_INT64, T_UINT64:
		return uint64(0)
	case T_FLOAT64:
		return float64(0)
	}

	return 0
}

func (n *NumericType) HasRange() bool { return !n.Range.IsEmpty() }

func (n *NumericType) GenerateHash(generator *HashGenerator) {
	generator.AddInt(int(n.dataType))
	generator.AddInt(int(n.Divisor))

	if n.Modulus != 0 {
		generator.AddInt(int(n.calculatedModulus.Uinteger))
	}

	if n.HasRange() {
		generator.AddInt(1)
		generator.AddInt(int(math.Floor(n.Range.Min.Float*float64(n.Divisor) + 0.5)))
		generator.AddInt(int(math.Floor(n.Range.Max.Float*float64(n.Divisor) + 0.5)))
	}
}
