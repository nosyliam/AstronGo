package dc

import (
	"errors"
	"fmt"
)

type Field interface {
	GetName() string
	SetName(name string) error

	GetType() BaseType
	SetType(ftype BaseType)

	HasDefaultValue() bool
	GetDefaultValue() interface{}
	SetDefaultValue(val interface{})

	GetStruct() Struct
	SetStruct(s Struct)
	GetId() uint
}

type AtomicField struct {
	Field
	KeywordList

	ftype        BaseType
	parentStruct *Struct

	id           uint
	name         string
	defaultValue *interface{}
}

func NewAtomicField(ftype BaseType, name string) Field {
	f := AtomicField{ftype: ftype, name: name}
	return f
}

func (f AtomicField) GetName() string { return f.name }
func (f AtomicField) SetName(name string) (err error) {
	if f.parentStruct != nil {
		if _, err := f.parentStruct.GetFieldByName(name); err != nil {
			return errors.New(fmt.Sprintf("field %s already exists in parent structure", name))
		}
	}

	f.name = name
	return nil
}

func (f AtomicField) GetType() BaseType      { return f.ftype }
func (f AtomicField) SetType(ftype BaseType) { f.ftype = ftype }

func (f AtomicField) HasDefaultValue() bool           { return f.defaultValue != nil }
func (f AtomicField) GetDefaultValue() interface{}    { return &f.defaultValue }
func (f AtomicField) SetDefaultValue(val interface{}) { f.defaultValue = &val }

func (f AtomicField) GetStruct() Struct  { return *f.parentStruct }
func (f AtomicField) SetStruct(s Struct) { f.parentStruct = &s }
func (f AtomicField) GetId() uint        { return f.id }
