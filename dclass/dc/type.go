package dc

// the Type constant defines the DistributedType's physical datatype. RealType
// defines a DistributedType's actual type class, e.g. NumericType or MethodType
type Type int
type RealType int

const (
	T_INVALID Type = iota
	T_INT8
	T_INT16
	T_INT32
	T_INT64
	T_UINT8
	T_UINT16
	T_UINT32
	T_UINT64
	T_CHAR
	T_FLOAT32
	T_FLOAT64

	T_STRING
	T_VARSTRING
	T_BLOB
	T_VARBLOB
	T_ARRAY
	T_VARARRAY

	T_STRUCT
	T_METHOD
)

const (
	INVALID_TYPE RealType = iota
	NUMERIC_TYPE
	ARRAY_TYPE
	STRUCT_TYPE
	METHOD_TYPE
)

type Sizetag_t uint32

type BaseType interface {
	GetType() Type

	HasRange() bool
	WithinRange(data []byte, length uint64) bool

	HasFixedSize() bool
	GetSize() Sizetag_t

	HasAlias() bool
	GetAlias() string
	SetAlias(string)

	GenerateHash(generator HashGenerator)
}

type DistributedType struct {
	dtype Type
	size  Sizetag_t
	alias string
}

func (d *DistributedType) GetType() Type                               { return d.dtype }
func (d *DistributedType) HasRange() bool                              { return false }
func (d *DistributedType) WithinRange(data []byte, length uint64) bool { return true }
func (d *DistributedType) HasFixedSize() bool                          { return d.size > 0 }
func (d *DistributedType) GetSize() Sizetag_t                          { return d.size }
func (d *DistributedType) HasAlias() bool                              { return len(d.alias) > 0 }
func (d *DistributedType) GetAlias() string                            { return d.alias }
func (d *DistributedType) SetAlias(alias string)                       { d.alias = alias }
func (d *DistributedType) GetDistributedType() RealType {
	switch (d.dtype) {
	case T_INT8:
	case T_INT16:
	case T_INT32:
	case T_INT64:
	case T_UINT8:
	case T_UINT16:
	case T_UINT32:
	case T_UINT64:
	case T_CHAR:
	case T_FLOAT32:
	case T_FLOAT64:
		return NUMERIC_TYPE
	case T_STRING:
	case T_VARSTRING:
	case T_BLOB:
	case T_VARBLOB:
	case T_ARRAY:
		return ARRAY_TYPE
	case T_STRUCT:
		return STRUCT_TYPE
	case T_METHOD:
		return METHOD_TYPE
	}
	return INVALID_TYPE
}
func (d *DistributedType) GenerateHash() {
	// todo
}
