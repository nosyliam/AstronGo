package main

import (
	"astrongo/core"
	"astrongo/dclass/dc"
	"fmt"
)

func main() {
	hasher := dc.NewHashGenerator()
	core.DC.GenerateHash(hasher)
	fmt.Printf("0x%x", hasher.Hash())
}
