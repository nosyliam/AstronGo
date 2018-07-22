package dc

import (
	"errors"
	"fmt"
)

type Field struct {
	KeywordList

	ftype BaseType
	parentStruct *Struct

	id uint
	name string
	defaultValue interface{}
}

func NewField(ftype BaseType, name string) Field {
	f := Field{ftype: ftype, name: name, defaultValue: ftype.GetDefaultValue()}
	return f
}

func (f Field) GetName() string { return f.name }
func (f Field) SetName(name string) (err error){
	if f.parentStruct != nil {
		if _, err := f.parentStruct.GetFieldByName(name); err != nil {
			return errors.New(fmt.Sprintf("field %s already exists in parent structure", name))
		}
	}

	f.name = name
	return nil
}

func (f Field) GetType() BaseType { return f.ftype }


func (f Field) HasDefaultValue() bool { return f.defaultValue != nil }
func (f Field) GetDefaultValue() interface{} { return f.defaultValue }