package shellquote

import (
	"testing"
	"testing/quick"
)

// this is called bothtest because it tests Split and Join together

func TestJoinSplit(t *testing.T) {
	f := func(strs []string) bool {
		// Join, then split, the input
		combined := Join(strs...)
		split, err := Split(combined)
		if err != nil {
			t.Logf("Error splitting %#v: %v", combined, err)
			return false
		}
		if !isEqualStrings(strs, split) {
			t.Logf("Input %q did not match output %q", strs, split)
			return false
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func isEqualStrings(s, t []string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := range s {
		if s[i] != t[i] {
			return false
		}
	}
	return true
}
