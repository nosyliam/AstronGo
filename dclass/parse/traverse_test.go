package parse

import (
	"astrongo/dclass/dc"
	"fmt"
	"testing"
)

func TestTraverse(t *testing.T) {
	dct, err := ParseFile("dclass/parse/test.dc")
	if err != nil {
		t.Fatalf("test dclass parse failed: %s", err)
	}

	dcf := dct.traverse()
	hashgen := dc.NewHashGenerator()
	dcf.GenerateHash(hashgen)
	fmt.Printf("Hash: %d", hashgen.Hash())
}
