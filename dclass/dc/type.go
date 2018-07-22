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

type Sizetag_t uint32

type BaseType interface {
	GetType() Type
	GetDefaultValue() interface{}

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
	BaseType
	dtype Type
	size  Sizetag_t
	alias string
}

func (d DistributedType) GetType() Type                               { return d.dtype }
func (d DistributedType) HasRange() bool                              { return false }
func (d DistributedType) WithinRange(data []byte, length uint64) bool { return true }
func (d DistributedType) HasFixedSize() bool                          { return d.size > 0 }
func (d DistributedType) GetSize() Sizetag_t                          { return d.size }
func (d DistributedType) HasAlias() bool                              { return len(d.alias) > 0 }
func (d DistributedType) GetAlias() string                            { return d.alias }
func (d DistributedType) SetAlias(alias string)                       { d.alias = alias }
func (d DistributedType) GetDefaultValue() interface{} {
	return ""
}

func (d DistributedType) GenerateHash(generator HashGenerator) {
	// todo
}
