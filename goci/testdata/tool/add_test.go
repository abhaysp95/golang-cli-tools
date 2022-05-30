package add

import "testing"

func TestAdd(t *testing.T) {
	a := 2
	b := 3
	res := 5

	if c := add(a, b); c != res {
		t.Errorf("Expected: %q, Got: %q.", res, c)
	}
}
