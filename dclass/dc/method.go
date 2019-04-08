package dc

import (
	"errors"
	"fmt"
)

type Method struct {
	DistributedType

	constrained bool

	parameters       []Parameter
	parametersByName map[string]Parameter
}

func NewMethod() *Method {
	m := &Method{}
	m.dataType = T_METHOD

	m.parameters = make([]Parameter, 0)
	m.parametersByName = make(map[string]Parameter, 0)
	return m
}

func (m Method) GetParameterByName(name string) (ok *Parameter, err error) {
	if val, ok := m.parametersByName[name]; ok {
		return &val, nil
	}
	return nil, errors.New(fmt.Sprintf("unable to index field `%s`", name))
}
func (m *Method) AddParameter(param Parameter) (err error) {
	if len(param.name) != 0 {
		if _, ok := m.parametersByName[param.name]; ok {
			return errors.New(fmt.Sprintf("parameter with name `%s` already exists", param.name))
		}

		m.parametersByName[param.name] = param
	}

	param.SetMethod(m)
	m.parameters = append(m.parameters, param)

	if m.HasFixedSize() || len(m.parameters) == 1 {
		if param.dataType.HasFixedSize() {
			m.size += param.dataType.Size()
		} else {
			m.size = 0
		}
	}
	return nil
}

func (m *Method) GenerateHash(generator HashGenerator) {
	// TODO
}
