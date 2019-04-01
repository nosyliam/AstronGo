package parse

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestTraverse(t *testing.T) {
	dc, err := ParseFile("dclass/parse/test.dc")
	if err != nil {
		t.Fatalf("test dclass parse failed: %s", err)
	}

	dcf := dc.traverse()
	spew.Dump(dcf)
}
