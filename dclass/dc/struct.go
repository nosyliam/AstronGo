package dc

import (
	"errors"
	"fmt"
)

type Struct struct {
	DistributedType
	file File

	id   uint
	name string

	fields       []Field
	fieldsByName map[string]Field
	fieldsById   map[uint]Field
}

func NewStruct(file File, name string) Struct {
	s := Struct{file: file, name: name}
	s.dtype = T_STRUCT
	return s
}

func (s Struct) GetFieldByName(name string) (ok *Field, err error) {
	if val, ok := s.fieldsByName[name]; ok {
		return &val, nil
	}
	return nil, errors.New(fmt.Sprintf("unable to index field `%s`", name))
}

func (s Struct) GetFieldById(id uint) (ok *Field, err error) {
	if val, ok := s.fieldsById[id]; ok {
		return &val, nil
	}
	return nil, errors.New(fmt.Sprintf("unable to index field id %d", id))
}

func (s Struct) AddField(field Field) (err error) {
	if fs := field.GetStruct(); &fs != &s {
		return errors.New("different structures cannot share the same field")
	}

	if _, ok := field.(MolecularField); ok {
		return errors.New("structures cannot contain molecular fields")
	}

	fieldName := field.GetName()
	if len(fieldName) == 0 {
		if fieldName == s.GetName() {
			return errors.New("structures cannot have constructors")
		}

		if _, ok := s.fieldsByName[fieldName]; ok {
			return errors.New(fmt.Sprintf("field with name `%s` already exists", fieldName))
		}

		s.fieldsByName[fieldName] = field
	}

	// TODO: add field to file
	s.fieldsById[field.GetId()] = field
	s.fields = append(s.fields, field)

	if s.HasFixedSize() || len(s.fields) == 1 {
		if field.GetType().HasFixedSize() {
			s.size += field.GetType().GetSize()
		} else {
			s.size = 0
		}
	}
	return nil
}

func (s Struct) GetName() string { return s.name }
func (s Struct) GetId() uint     { return s.id }
func (s Struct) SetId(id uint)   { s.id = id }

func (s Struct) GenerateHash(generator HashGenerator) {

}
