package core

import (
	"astrongo/dclass/dc"
	"astrongo/dclass/parse"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

var (
	dcf    *dc.File
	dcOnce sync.Once
)

func LoadDC() *dc.File {
	var configs []string
	config := GetConfig()

	for _, conf := range config.General.DCFiles {
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
	dcf = dctree.Traverse()
	return dcf
}

func GetDC() *dc.File {
	dcOnce.Do(func() {
		dcf = LoadDC()
	})

	if dcf == nil {
		os.Exit(1)
	}

	return dcf
}
