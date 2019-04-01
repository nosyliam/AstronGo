package parse

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	dc, err := ParseFile("dclass/parse/test.dc")
	if err != nil {
		t.Fatalf("test dclass parse failed: %s", err)
	}

	fmt.Print(dc)
}
