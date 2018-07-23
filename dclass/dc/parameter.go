package dc

import (
	"errors"
	"fmt"
)

type Parameter struct {
	method *Method
	ptype BaseType

	name string
	typeAlias string
	defaultValue *interface{}
}

func NewParameter(ptype BaseType, name string) Parameter {
	p := Parameter{ptype: ptype, name: name}
	return p
}


func (p Parameter) GetName() string { return p.name }
func (p Parameter) SetName(name string) (err error) {
	if p.method != nil {
		if _, err := p.method.GetParameterByName(name); err != nil {
			return errors.New(fmt.Sprintf("parameter %s already exists in parent method", name))
		}
	}

	p.name = name
	return nil
}

func (p Parameter) GetType() BaseType      { return p.ptype }
func (p Parameter) SetType(ftype BaseType) (err error) {
	if ftype.GetType() == T_METHOD {
		return errors.New("parameters cannot have method types")
	}

	if _, ok := ftype.(Class); ok {
		return errors.New("parameters cannot have class types")
	}

	p.ptype = ftype
	p.defaultValue = nil
	return nil
}

func (p Parameter) HasDefaultValue() bool           { return p.defaultValue != nil }
func (p Parameter) GetDefaultValue() interface{}    { return &p.defaultValue }
func (p Parameter) SetDefaultValue(val interface{}) { p.defaultValue = &val }

func (p Parameter) GetMethod() Method { return *p.method }
func (p Parameter) SetMethod(method Method) { p.method = &method }
