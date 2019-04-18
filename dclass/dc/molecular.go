package dc

type MolecularField struct {
	BaseField
	Struct
}

func NewMolecular() *MolecularField {
	m := &MolecularField{}
	m.fieldType = BaseType(m)
	return m
}

func (s *MolecularField) AddField(field Field) (err error) {
	return nil
}
