package dc

import "errors"

type MolecularField struct {
	BaseField
	Struct
}

func NewMolecularField(name string) *MolecularField {
	m := &MolecularField{}
	m.BaseField.name = name
	m.BaseField.fieldType = BaseType(m)
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
