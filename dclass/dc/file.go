package dc

import (
	"errors"
	"fmt"
)

// get_class
// get_struct
// get_class_by_id
// get_class_by_name
// get_num_types
// get_type_by_id
// get_type_by_name
// get_field_by_id
// get_num_imports
// get_import
// has_keyword
// get_num_keyword
// get_keyword
// add_class
// add_struct
// add_import
// add_typedef
// add_keyword
// get_hash

type File struct {
	KeywordList

	classes  []*Class
	structs  []*Struct
	typedefs map[string]BaseType

	fields      []*Field
	types       []BaseType
	typesByName map[string]BaseType
}

func NewFile() *File {
	f := File{}
	f.typedefs = make(map[string]BaseType, 0)
	f.typesByName = make(map[string]BaseType, 0)

	return &f
}

func (f File) TypeById(n int) (t *BaseType, ok bool) {
	if n >= len(f.types) {
		return nil, false
	}

	return &f.types[n], true
}

func (f File) TypeByName(name string) (t *BaseType, ok bool) {
	if t, ok := f.typedefs[name]; ok {
		return &t, true
	}

	if t, ok := f.typesByName[name]; ok {
		return &t, true
	}

	return nil, false
}

func (f *File) AddTypedef(name string, t BaseType) (err error) {
	if _, ok := f.typedefs[name]; ok {
		return errors.New(fmt.Sprint("typedef `%s` is already declared", name))
	}

	f.typedefs[name] = t
	return nil
}

func (f File) Class(n int) (t *Class, ok bool) {
	if n >= len(f.classes) {
		return nil, false
	}

	return f.classes[n], ok
}

func (f *File) AddClass(class *Class) (err error) {
	if _, ok := f.TypeByName(class.name); ok {
		return errors.New(fmt.Sprint("type `%s` is already defined", class.name))
	}

	class.id = uint(len(f.types))
	f.types = append(f.types, BaseType(class))
	f.classes = append(f.classes, class)
	return nil
}

func (f File) ClassByName(name string) (class *Class, ok bool) {
	t, ok := f.TypeByName(name)
	if !ok {
		return nil, false
	}

	if class, ok := (*t).(*Class); !ok {
		return nil, false
	} else {
		return class, true
	}
}

func (f File) Struct(n int) (t *Struct, ok bool) {
	if n >= len(f.classes) {
		return nil, false
	}

	return f.structs[n], true
}

func (f *File) AddStruct(strct *Struct) (err error) {
	if _, ok := f.TypeByName(strct.name); ok {
		return errors.New(fmt.Sprint("type `%s` is already defined", strct.name))
	}

	strct.id = uint(len(f.types))
	f.types = append(f.types, BaseType(strct))
	f.structs = append(f.structs, strct)
	return nil
}

func (f File) Field(n int) (t *Field, ok bool) {
	if n >= len(f.fields) {
		return nil, false
	}

	return f.fields[n], true
}

func (f *File) AddField(field *Field) /* cannot throw an error */ {
	(*field).SetId(uint(len(f.fields)))
	f.fields = append(f.fields, field)
}

func (f File) GenerateHash(generator HashGenerator) {
	// TODO
}
