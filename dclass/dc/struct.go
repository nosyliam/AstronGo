package dc

import (
	"errors"
	"fmt"
)

type Struct struct {
	DistributedType

	id          uint
	name        string
	constrained bool

	fieldsByName map[string]Field
	fieldsById   map[uint]Field
	fields       []Field

	file *File
}

func NewStruct(name string, file *File) *Struct {
	s := &Struct{name: name, file: file}
	s.dataType = T_STRUCT

	s.fieldsByName = make(map[string]Field, 0)
	s.fieldsById = make(map[uint]Field, 0)
	return s
}

func (s Struct) GetFieldByName(name string) (field *Field, ok bool) {
	if val, ok := s.fieldsByName[name]; ok {
		return &val, true
	}
	return nil, false
}

func (s Struct) GetFieldById(id uint) (field *Field, ok bool) {
	if val, ok := s.fieldsById[id]; ok {
		return &val, true
	}
	return nil, false
}

func (s *Struct) AddField(field Field) (err error) {
	if _, ok := field.(*MolecularField); ok {
		return errors.New("structures cannot contain methods")
	}

	if _, ok := field.(*MolecularField); ok {
		return errors.New("structures cannot contain molecular fields")
	}

	fieldName := field.Name()
	if len(fieldName) == 0 {
		if fieldName == s.name {
			return errors.New("structures cannot have constructors")
		}

		if _, ok := s.fieldsByName[fieldName]; ok {
			return errors.New(fmt.Sprintf("field with name `%s` already exists", fieldName))
		}

		s.fieldsByName[fieldName] = field
	}

	field.SetParentStruct(s)
	s.file.AddField(&field)
	s.fieldsById[field.Id()] = field
	s.fields = append(s.fields, field)

	if s.HasFixedSize() || len(s.fields) == 1 {
		if field.FieldType().HasFixedSize() {
			s.size += field.FieldType().Size()
		} else {
			s.size = 0
		}
	}

	s.constrained = field.FieldType().HasRange()
	return nil
}

func (s *Struct) GenerateHash(generator *HashGenerator) {
	generator.AddString(s.name)
	generator.AddInt(1)
	generator.AddInt(0)

	generator.AddInt(len(s.fields))
	for _, field := range s.fields {
		field.GenerateHash(generator)
	}
}
