package dc

import (
	"errors"
	"fmt"
)

type Class struct {
	Struct
	file File

	baseFields map[string]Field
}

func NewClass(file File, name string) Class {
	c := Class{file: file}
	c.dataType = T_STRUCT
	c.name = name

	c.baseFields = make(map[string]Field, 0)
	return c
}

func (c Class) AddParent(class Class) {

}

func (c Class) AddField(field Field) (err error) {
	if fs := field.Struct(); &fs != &c.Struct {
		return errors.New("different classes cannot share the same field")
	}

	if len(field.Name()) == 0 {
		return errors.New("class field names cannot be empty")
	}

	fieldName := field.Name()
	if fieldName == c.name {
		if _, ok := field.(MolecularField); ok {
			return errors.New("constructors cannot be molecular fields")
		}

		if len(c.baseFields) > 0 {
			return errors.New("constructor must be the first field in a class")
		}
	}

	if _, ok := c.baseFields[fieldName]; ok {
		return errors.New(fmt.Sprintf("field with name `%s` already exists", fieldName))
	}

	// TODO: add to file
	c.fieldsById[field.Id()] = field
	c.fieldsByName[fieldName] = field
	c.baseFields[fieldName] = field



	return nil
}

