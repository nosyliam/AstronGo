package main

import (
	"astrongo/core"
	"astrongo/dclass/dc"
	"github.com/apex/log"
)

var mainLog *log.Entry

func init() {
	log.SetHandler(core.Log)
	log.SetLevel(log.DebugLevel)
	mainLog = log.WithFields(log.Fields{
		"name": "Main",
	})
}

func main() {

	hasher := dc.NewHashGenerator()
	core.DC.GenerateHash(hasher)
}
