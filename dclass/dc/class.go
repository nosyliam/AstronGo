package dc

import "errors"

type Class struct {
	Struct
	file File

	constructor      Field
	baseFields       []Field
	baseFieldsByName map[string]Field

	parents  []Class
	children []Class
}

func NewClass(file File, name string) Class {
	c := Class{file: file}
	c.dataType = T_STRUCT
	c.Name = name

	c.baseFields = make([]Field, 0)
	c.baseFieldsByName = make(map[string]Field, 0)
	c.parents = make([]Class, 0)
	c.children = make([]Class, 0)
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

	if field.Name() == c.Name {
		if _, ok := field.(MolecularField); ok {
			return errors.New("constructors cannot be molecular fields")
		}

		if len(c.baseFields) > 0 {
			return errors.New("constructor must be the first field in a class")
		}
	}

	return nil
}

func (c Class) addChild(class Class) {

}

func (c Class) addInheritedField(parent Class, field Field) {

}

func (c Class) shadowField(field Field) {

}

