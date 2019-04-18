package dc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Field interface {
	SetName(string) error

	FieldType() BaseType
	Name() string
	Id() uint
	Keywords() KeywordList
	SetId(uint)

	SetParentStruct(*Struct)
	ParentStruct() *Struct

	HasDefaultValue() bool
	SetDefaultValue([]interface{})
	FieldDefaultValue() []byte
}

type BaseField struct {
	Field
	KeywordList

	fieldType BaseType
	id        uint
	name      string

	defaultValue []interface{}
	parentStruct *Struct
}

func (f *BaseField) FieldType() BaseType   { return f.fieldType }
func (f *BaseField) Name() string          { return f.name }
func (f *BaseField) Id() uint              { return f.id }
func (f *BaseField) Keywords() KeywordList { return f.KeywordList }
func (f *BaseField) SetId(id uint)         { f.id = id }

func (f *BaseField) SetParentStruct(s *Struct) { f.parentStruct = s }
func (f *BaseField) ParentStruct() *Struct     { return f.parentStruct }

func (f *BaseField) HasDefaultValue() bool              { return f.defaultValue != nil }
func (f *BaseField) SetDefaultValue(data []interface{}) { f.defaultValue = data }
func (f *BaseField) FieldDefaultValue() []byte {
	if f.defaultValue == nil {
		f.defaultValue = append(make([]interface{}, 0), f.fieldType.DefaultValue())
	}

	buf := new(bytes.Buffer)
	for _, v := range f.defaultValue {
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			panic(fmt.Sprintf("unable to decode field default value: %s", err))
		}
	}

	return buf.Bytes()
}

type AtomicField struct {
	BaseField
}

func NewAtomicField(dataType BaseType, name string) Field {
	f := &AtomicField{BaseField{fieldType: dataType, name: name}}
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
