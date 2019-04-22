package parse

import (
	"astrongo/dclass/dc"
	"testing"
)

const DC_HASH = 0xb77baf2

func TestTraverse(t *testing.T) {
	dct, err := ParseFile("dclass/parse/test.dc")
	if err != nil {
		t.Fatalf("test dclass parse failed: %s", err)
	}

	dcf := dct.Traverse()
	hashgen := dc.NewHashGenerator()
	dcf.GenerateHash(hashgen)
	hash := hashgen.Hash()
	if hash != DC_HASH {
		t.Fatalf("test dclass dc mismatch: 0x%x", hash)
	}
}
