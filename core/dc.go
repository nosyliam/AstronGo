package core

import (
	"astrongo/dclass/dc"
	"astrongo/dclass/parse"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var DC *dc.File

func loadDC() *dc.File {
	var configs []string

	for _, conf := range Config.General.DC_Files {
		fmt.Println(conf)
		data, err := ioutil.ReadFile(conf)
		if err != nil {
			fmt.Printf("failed to read dc file %s: %v", conf, err)
			return nil
		}

		configs = append(configs, string(data))
	}

	dctree, err := parse.ParseString(strings.Join(configs, "\n"))
	if err != nil {
		fmt.Printf("error while parsing dc: %v", err)
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error while traversing dc: ", r)
			os.Exit(1)
		}
	}()
	return dctree.Traverse()
}

func init() {
	DC = loadDC()
	if DC == nil {
		os.Exit(1)
	}
}
