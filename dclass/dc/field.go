package dc

import (
	"errors"
	"fmt"
)

type Field interface {
	SetName(string) error

	FieldType() BaseType
	Name() string
	Id() uint
	SetId(uint)

	SetParentStruct(*Struct)
	ParentStruct() *Struct

	HasDefaultValue() bool
	SetDefaultValue([]interface{})
	FieldDefaultValue() []interface{}
}

type AtomicField struct {
	Field
	KeywordList

	fieldType BaseType
	id        uint
	name      string

	defaultValue []interface{}
	parentStruct *Struct
}

func NewAtomicField(dataType BaseType, name string) Field {
	f := &AtomicField{fieldType: dataType, name: name}
	f.keywords = make(map[string]struct{}, 0)
	return f
}

func (f *AtomicField) SetName(name string) (err error) {
	if f.parentStruct != nil {
		if _, ok := f.parentStruct.GetFieldByName(name); ok {
			return errors.New(fmt.Sprintf("field %s already exists in parent struct", name))
		}
	}

	f.name = name
	return nil
}

func (f *AtomicField) FieldType() BaseType { return f.fieldType }
func (f *AtomicField) Name() string        { return f.name }
func (f *AtomicField) Id() uint            { return f.id }
func (f *AtomicField) SetId(id uint)       { f.id = id }

func (f *AtomicField) SetParentStruct(s *Struct) { f.parentStruct = s }
func (f *AtomicField) ParentStruct() *Struct     { return f.parentStruct }

func (f *AtomicField) HasDefaultValue() bool              { return f.defaultValue != nil }
func (f *AtomicField) SetDefaultValue(data []interface{}) { f.defaultValue = data }
func (f *AtomicField) FieldDefaultValue() []interface{} {
	if f.HasDefaultValue() {
		return f.defaultValue
	}
	return append(make([]interface{}, 0), f.fieldType.DefaultValue())
}
