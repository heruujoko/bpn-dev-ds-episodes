package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestMenu(t *testing.T) {
	for _, input := range []string{"", "0", "bad\n9\n0\n", "1", "1\n\n0\n"} {
		var out bytes.Buffer
		menu(strings.NewReader(input), &out)
		if strings.Contains(input, "bad") && (!strings.Contains(out.String(), "Invalid choice \"bad\"") || !strings.Contains(out.String(), "Invalid choice \"9\"")) {
			t.Fatal(out.String())
		}
		if strings.HasPrefix(input, "1") && !strings.Contains(out.String(), "Dependency calls: A=4 B=4 C=4") {
			t.Fatal(out.String())
		}
	}
}
