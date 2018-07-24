package dc

import (
	"errors"
	"fmt"
)

type Parameter struct {
	Type      BaseType
	Name      string
	TypeAlias string

	defaultValue interface{}
	method       *Method
}

func NewParameter(dataType BaseType, name string) Parameter {
	p := Parameter{Type: dataType, Name: name}
	return p
}

func (p Parameter) SetName(name string) (err error) {
	if p.method != nil {
		if _, err := p.method.GetParameterByName(name); err != nil {
			return errors.New(fmt.Sprintf("parameter %s already exists in parent method", name))
		}
	}

	p.Name = name
	return nil
}

func (p Parameter) SetType(dataType BaseType) (err error) {
	if dataType.Type() == T_METHOD {
		return errors.New("parameters cannot have method types")
	}

	if _, ok := dataType.(Class); ok {
		return errors.New("parameters cannot have class types")
	}

	p.Type = dataType
	p.defaultValue = nil
	return nil
}

func (p Parameter) SetMethod(method Method) { p.method = &method }
func (p Parameter) Method() Method          { return *p.method }

func (p Parameter) HasDefaultValue() bool           { return p.defaultValue != nil }
func (p Parameter) SetDefaultValue(val interface{}) { p.defaultValue = &val }
func (p Parameter) GetDefaultValue() interface{}    { return &p.defaultValue }
