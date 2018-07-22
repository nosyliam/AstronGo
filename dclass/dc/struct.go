package dc

import (
	"fmt"
	"errors"
)

type Struct struct {
	DistributedType
	file File

	id uint
	name string

	fields []Field
	fieldsByName map[string]Field
	fieldsById map[uint]Field
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
	return nil, errors.New(fmt.Sprintf("unable to index field %s", name))
}

func (s Struct) GetFieldById(id uint) (ok *Field, err error) {
	if val, ok := s.fieldsById[id]; ok {
		return &val, nil
	}
	return nil, errors.New(fmt.Sprintf("unable to index field id %d", id))
}

func (s Struct) GenerateHash(generator HashGenerator) {

}