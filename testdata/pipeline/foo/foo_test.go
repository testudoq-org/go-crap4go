// This file should be excluded from analysis (test files are skipped).
package foo

import "testing"

func TestSimple(t *testing.T) {
	if Simple() != 1 {
		t.Error("want 1")
	}
}
