package main

import (
	"astrongo/core"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	dcf := core.GetDC()
	spew.Dump(dcf)
}
