package dc

type Type int

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

func StringToType(s string) Type {
	switch s {
	case "int8":
		return T_INT8
	case "int16":
		return T_INT16
	case "int32":
		return T_INT32
	case "int64":
		return T_INT64
	case "uint8":
		return T_UINT8
	case "uint16":
		return T_UINT16
	case "uint32":
		return T_UINT32
	case "uint64":
		return T_UINT64
	case "char":
		return T_CHAR
	case "float":
	case "float32":
		return T_FLOAT32
	case "float64":
		return T_FLOAT64
	case "string":
		return T_STRING
	case "blob":
		return T_BLOB
	case "array":
		return T_ARRAY
	}

	return T_INVALID
}

type Sizetag_t uint32

type BaseType interface {
	Type() Type
	DefaultValue() interface{}
	SetAlias(string)

	HasRange() bool
	WithinRange([]byte, uint64) bool

	HasFixedSize() bool
	Size() Sizetag_t

	GenerateHash(HashGenerator)
}

type DistributedType struct {
	BaseType

	dataType Type
	size     Sizetag_t
	alias    string
}

func (d DistributedType) Type() Type                                  { return d.dataType }
func (d DistributedType) HasRange() bool                              { return false }
func (d DistributedType) SetAlias(alias string)                       { d.alias = alias }
func (d DistributedType) WithinRange(data []byte, length uint64) bool { return true }
func (d DistributedType) HasFixedSize() bool                          { return d.size > 0 }
func (d DistributedType) Size() Sizetag_t                             { return d.size }
func (d DistributedType) DefaultValue() interface{} {
	return ""
}

func (d DistributedType) GenerateHash(generator HashGenerator) {
	// todo
}
