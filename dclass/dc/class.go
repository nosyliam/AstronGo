package dc

import (
	"errors"
	"fmt"
	"sort"
)

type Class struct {
	Struct

	file *File

	baseFields  []Field
	constructor Field
	parents     []Class
}

func NewClass(name string, file *File) *Class {
	c := &Class{file: file}
	c.dataType = T_STRUCT
	c.name = name

	c.fieldsByName = make(map[string]Field, 0)
	c.fieldsById = make(map[uint]Field, 0)
	return c
}

func (c *Class) HasField(name string) bool {
	for _, field := range c.baseFields {
		if field.Name() == name {
			return true
		}
	}
	return false
}

func (c *Class) AddParent(class Class) {
	c.parents = append(c.parents, class)

	for _, field := range class.baseFields {
		c.AddInheritedField(field)
	}
}

func (c *Class) AddField(field Field) (err error) {
	if len(field.Name()) == 0 {
		return errors.New("class field names cannot be empty")
	}

	fieldName := field.Name()
	if fieldName == c.name {
		if c.constructor != nil {
			return errors.New("class already has a constructor")
		}

		if _, ok := field.(*MolecularField); ok {
			return errors.New("constructors cannot be molecular fields")
		}

		if len(c.baseFields) > 0 {
			return errors.New("constructor must be the first field in a class")
		}

		c.file.AddField(&field)
		c.constructor = field

		c.fieldsById[field.Id()] = field
		c.fieldsByName[fieldName] = field
		c.baseFields = append(c.baseFields, field)
		return nil
	}

	if c.HasField(fieldName) {
		return errors.New(fmt.Sprintf("field with name `%s` already exists", fieldName))
	}

	c.file.AddField(&field)
	c.fields = append(c.fields, field)

	c.fieldsById[field.Id()] = field
	c.fieldsByName[fieldName] = field
	c.baseFields = append(c.baseFields, field)

	if c.HasFixedSize() || len(c.fields) == 1 {
		if field.FieldType().HasFixedSize() {
			c.size += field.FieldType().Size()
		} else {
			c.size = 0
		}
	}

	c.constrained = c.constrained || field.FieldType().HasRange()
	return nil
}

func (c *Class) AddInheritedField(field Field) {
	fieldName := field.Name()
	if c.HasField(fieldName) {
		return
	}

	c.fieldsById[field.Id()] = field
	c.fieldsByName[fieldName] = field

	c.fields = append(c.fields, field)
	sort.Slice(c.fields, func(i, j int) bool {
		return c.fields[i].Id() < c.fields[j].Id()
	})

	if c.HasFixedSize() || len(c.fields) == 1 {
		if field.FieldType().HasFixedSize() {
			c.size += field.FieldType().Size()
		} else {
			c.size = 0
		}
	}

	c.constrained = c.constrained || field.FieldType().HasRange()
}

func (c *Class) HasRange() bool {
	return c.constrained
}

func (c *Class) Name() string {
	return c.name
}

func (c *Class) GenerateHash(generator *HashGenerator) {
	generator.AddString(c.name)

	generator.AddInt(len(c.parents))
	for _, parent := range c.parents {
		generator.AddInt(int(parent.id))
	}

	if c.constructor != nil {
		c.constructor.GenerateHash(generator)
	}

	generator.AddInt(len(c.baseFields))
	for _, field := range c.baseFields {
		field.GenerateHash(generator)
	}
}
