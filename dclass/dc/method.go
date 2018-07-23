package dc

import (
	"fmt"
	"errors"
)

type Method struct {
	DistributedType

	parameters []Parameter
	parametersByName map[string]Parameter
}

func NewMethod() Method {
	m := Method{}
	m.dtype = T_METHOD
	return m
}

func (m Method) GetNumParameters() int { return len(m.parameters) }
func (m Method) GetParameter(n int) Parameter { return m.parameters[n] }
func (m Method) GetParameterByName(name string) (ok *Parameter, err error) {
	if val, ok := m.parametersByName[name]; ok {
		return &val, nil
	}
	return nil, errors.New(fmt.Sprintf("unable to index field `%s`", name))
}
func (m Method) AddParameter(param Parameter) (err error) {
	paramName := param.GetName()
	if len(paramName) == 0 {
		if _, ok := m.parametersByName[paramName]; ok {
			return errors.New(fmt.Sprintf("parameter with name `%s` already exists", paramName))
		}

		m.parametersByName[paramName] = param
	}

	param.SetMethod(m)
	m.parameters = append(m.parameters, param)

	if m.HasFixedSize() || len(m.parameters) == 1 {
		if param.GetType().HasFixedSize() {
			m.size += param.GetType().GetSize()
		} else {
			m.size = 0
		}
	}
	return nil
}

func (m Method) GenerateHash(generator HashGenerator) {
	// TODO
}