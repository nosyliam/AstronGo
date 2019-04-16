package dc

type MolecularField struct {
	Field
	Struct

	fieldType BaseType
	id        uint
	name      string

	defaultValue []interface{}
	parentStruct *Struct
}

func NewMolecular() *MolecularField {
	m := &MolecularField{}
	m.fieldType = BaseType(m)
	return m
}

func (s *MolecularField) AddField(field Field) (err error) {
	return nil
}

func (f *MolecularField) FieldType() BaseType { return f.fieldType }
func (f *MolecularField) Name() string        { return f.name }
func (f *MolecularField) Id() uint            { return f.id }
func (f *MolecularField) SetId(id uint)       { f.id = id }

func (f *MolecularField) SetParentStruct(s *Struct) { f.parentStruct = s }
func (f *MolecularField) ParentStruct() *Struct     { return f.parentStruct }

func (f *MolecularField) HasDefaultValue() bool { return f.defaultValue != nil }
func (f *MolecularField) FieldDefaultValue() []interface{} {
	if f.HasDefaultValue() {
		return f.defaultValue
	}
	return append(make([]interface{}, 0), f.fieldType.DefaultValue())
}
