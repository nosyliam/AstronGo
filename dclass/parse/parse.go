package parse

import (
	"github.com/alecthomas/participle"
	"io/ioutil"
)

func ParseFile(file string) (ok *DCFile, err error) {
	parser, err := participle.Build(&DCFile{}, participle.UseLookahead(16))
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	dc := &DCFile{}
	err = parser.ParseString(string(b), dc)
	if err != nil {
		return nil, err
	}

	return dc, nil
}

func ParseString(conf string) (ok *DCFile, err error) {
	parser, err := participle.Build(&DCFile{}, participle.UseLookahead(16))
	if err != nil {
		return nil, err
	}

	dc := &DCFile{}
	err = parser.ParseString(conf, dc)
	if err != nil {
		return nil, err
	}

	return dc, nil
}
