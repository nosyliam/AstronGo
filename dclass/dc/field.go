package dc

import (
	"errors"
	"fmt"
)

type Field interface {
	SetName(string) error

	Type() BaseType
	Name() string
	Id() uint
	SetId(uint)

	SetStruct(Struct)
	Struct() Struct

	HasDefaultValue() bool
	DefaultValue() interface{}
}

type AtomicField struct {
	Field
	KeywordList

	fieldType BaseType
	id        uint
	name      string

	defaultValue *interface{}
	parentStruct *Struct
}

func NewAtomicField(dataType BaseType, name string) Field {
	f := AtomicField{fieldType: dataType, name: name}
	f.keywords = make(map[string]struct{}, 0)
	return f
}

func (f AtomicField) SetName(name string) (err error) {
	if f.parentStruct != nil {
		if _, ok := f.parentStruct.GetFieldByName(name); ok {
			return errors.New(fmt.Sprintf("field %s already exists in parent struct", name))
		}
	}

	f.name = name
	return nil
}

func (f AtomicField) Type() BaseType { return f.fieldType }
func (f AtomicField) Name() string   { return f.name }
func (f AtomicField) Id() uint       { return f.id }
func (f AtomicField) SetId(id uint)  { f.id = id }

func (f AtomicField) SetStruct(s Struct) { f.parentStruct = &s }
func (f AtomicField) Struct() Struct     { return *f.parentStruct }

func (f AtomicField) HasDefaultValue() bool { return f.defaultValue != nil }
func (f AtomicField) DefaultValue() interface{} {
	if f.HasDefaultValue() {
		return *f.defaultValue
	}
	return f.fieldType.DefaultValue()
}
