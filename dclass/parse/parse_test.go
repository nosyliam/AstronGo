package parse

import (
	"testing"
)

func TestParse(t *testing.T) {
	if _, err := ParseFile("dclass/parse/test.dc"); err != nil {
		t.Fatalf("test dclass parse failed: %s", err)
	}
}
