package core

import (
	"astrongo/dclass/dc"
	"astrongo/dclass/parse"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

var DC *dc.File

func LoadDC() (err error) {
	var configs []string

	for _, conf := range Config.General.DC_Files {
		data, err := ioutil.ReadFile(conf)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to read DC file %s: %v", conf, err))
		}

		configs = append(configs, string(data))
	}

	dctree, err := parse.ParseString(strings.Join(configs, "\n"))
	if err != nil {
		return errors.New(fmt.Sprintf("Error while parsing DC file: %v", err))
	}

	err = nil
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("Error while traversing DC file: %v", r))
		}
	}()

	DC = dctree.Traverse()
	return err
}
