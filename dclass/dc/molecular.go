package dc

import "errors"

type MolecularField struct {
	BaseField
	Struct
}

func NewMolecularField(name string) *MolecularField {
	m := &MolecularField{}
	m.fieldType = BaseType(m)
	m.BaseField.name = name
	return m
}

func (m *MolecularField) AddField(field Field) (err error) {
	if _, ok := field.(*MolecularField); ok {
		return errors.New("molecular fields cannot be nested")
	}

	if len(m.fields) == 0 {
		m.Copy(field.Keywords())
	} else if !m.HasMatchingKeywords(field.Keywords()) {
		return errors.New("fields in a molecular must have matching keywords")
	}

	m.fields = append(m.fields, field)

	if m.HasFixedSize() || len(m.fields) == 1 {
		if field.FieldType().HasFixedSize() {
			m.size += field.FieldType().Size()
		} else {
			m.size = 0
		}
	}

	return nil
}

func (m *MolecularField) GenerateHash(generator *HashGenerator) {
	generator.AddString(m.BaseField.name)
	generator.AddInt(int(m.BaseField.id))

	generator.AddInt(len(m.fields))
	for _, field := range m.fields {
		field.GenerateHash(generator)
	}
}
