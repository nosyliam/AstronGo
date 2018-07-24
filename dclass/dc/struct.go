package dc

import (
	"errors"
	"fmt"
)

type Struct struct {
	DistributedType

	Id   uint
	Name string

	fields       []Field
	fieldsByName map[string]Field
	fieldsById   map[uint]Field
}

func NewStruct(name string) Struct {
	s := Struct{Name: name}
	s.dataType = T_STRUCT
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
	if fs := field.Struct(); &fs != &s {
		return errors.New("different structures cannot share the same field")
	}

	if _, ok := field.(MolecularField); ok {
		return errors.New("structures cannot contain molecular fields")
	}

	fieldName := field.Name()
	if len(fieldName) == 0 {
		if fieldName == s.Name {
			return errors.New("structures cannot have constructors")
		}

		if _, ok := s.fieldsByName[fieldName]; ok {
			return errors.New(fmt.Sprintf("field with name `%s` already exists", fieldName))
		}

		s.fieldsByName[fieldName] = field
	}

	// TODO: add field to file
	s.fieldsById[field.Id()] = field
	s.fields = append(s.fields, field)

	if s.HasFixedSize() || len(s.fields) == 1 {
		if field.Type().HasFixedSize() {
			s.size += field.Type().Size()
		} else {
			s.size = 0
		}
	}
	return nil
}

func (s Struct) GenerateHash(generator HashGenerator) {

}
