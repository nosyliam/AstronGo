package dc

import (
	"errors"
	"fmt"
)

type Method struct {
	DistributedType

	parameters       []Parameter
	parametersByName map[string]Parameter
}

func NewMethod() Method {
	m := Method{}
	m.dataType = T_METHOD
	return m
}

func (m Method) GetParameterByName(name string) (ok *Parameter, err error) {
	if val, ok := m.parametersByName[name]; ok {
		return &val, nil
	}
	return nil, errors.New(fmt.Sprintf("unable to index field `%s`", name))
}
func (m Method) AddParameter(param Parameter) (err error) {
	if len(param.Name) == 0 {
		if _, ok := m.parametersByName[param.Name]; ok {
			return errors.New(fmt.Sprintf("parameter with name `%s` already exists", param.Name))
		}

		m.parametersByName[param.Name] = param
	}

	param.SetMethod(m)
	m.parameters = append(m.parameters, param)

	if m.HasFixedSize() || len(m.parameters) == 1 {
		if param.Type.HasFixedSize() {
			m.size += param.Type.Size()
		} else {
			m.size = 0
		}
	}
	return nil
}

func (m Method) GenerateHash(generator HashGenerator) {
	// TODO
}
