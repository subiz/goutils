package expression

import (
	"fmt"
	"testing"
)

func TestParseAndEval(t *testing.T) {
	var tests = []struct {
		exp string
		res bool
	}{
		{"(true OR false) AND false", false},
		{"true OR false AND false", true},
	}

	for _, v := range tests {
		if r, err := ParseAndEval(v.exp); err == nil {
			fmt.Println(v.exp, "=", r)
			if r != v.res {
				t.Errorf("expect %v, got %v in '%s'", v.res, r, v.exp)
			}
		} else {
			t.Fatalf("%s: %v\n", v.exp, err)
		}
	}
}
