package dc

import (
	"errors"
	"fmt"
	"sort"
)

type Class struct {
	Struct

	file File

	baseFields map[string]Field
	parents    []Class
}

func NewClass(file File, name string) *Class {
	c := &Class{file: file}
	c.dataType = T_STRUCT
	c.name = name

	c.baseFields = make(map[string]Field, 0)
	return c
}

func (c *Class) AddParent(class Class) {
	c.parents = append(c.parents, class)

	for _, field := range class.baseFields {
		c.AddInheritedField(field)
	}
}

func (c *Class) AddField(field Field) (err error) {
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

	c.file.AddField(&field)
	c.fields = append(c.fields, field)

	c.fieldsById[field.Id()] = field
	c.fieldsByName[fieldName] = field
	c.baseFields[fieldName] = field

	if c.HasFixedSize() || len(c.fields) == 1 {
		if field.Type().HasFixedSize() {
			c.size += field.Type().Size()
		} else {
			c.size = 0
		}
	}

	return nil
}

func (c *Class) AddInheritedField(field Field) {
	fieldName := field.Name()
	if _, ok := c.baseFields[fieldName]; ok {
		return
	}

	c.fieldsById[field.Id()] = field
	c.fieldsByName[fieldName] = field

	c.fields = append(c.fields, field)
	sort.Slice(c.fields, func(i, j int) bool {
		return c.fields[i].Id() < c.fields[j].Id()
	})

	if c.HasFixedSize() || len(c.fields) == 1 {
		if field.Type().HasFixedSize() {
			c.size += field.Type().Size()
		} else {
			c.size = 0
		}
	}

	c.constrained = field.Type().HasRange()
}

func (c *Class) GenerateHash(generator HashGenerator) {
	// TODO
}
