package coverage_test

import (
	"testing"

	"github.com/go-crap4go/crap4go/internal/coverage"
)

// TestLoadProfile_NonExistentFile verifies that LoadProfile returns an error
// when the given file does not exist.
func TestLoadProfile_NonExistentFile(t *testing.T) {
	t.Skip("TODO(prompt-3): implement coverage profile parsing")
	_, err := coverage.LoadProfile("testdata/nonexistent.out")
	if err == nil {
		t.Error("LoadProfile returned nil error for non-existent file; want error")
	}
}

// TestLoadProfile_ValidProfile will verify that a well-formed coverage.out
// file is parsed into a Profile with the correct mode and block count.
func TestLoadProfile_ValidProfile(t *testing.T) {
	t.Skip("TODO(prompt-3): implement coverage profile parsing")
}

// TestLoadProfile_EmptyFile will verify that an empty file returns an error.
func TestLoadProfile_EmptyFile(t *testing.T) {
	t.Skip("TODO(prompt-3): implement coverage profile parsing")
}

// TestMapCoverage_ReturnsNonNilMap verifies that MapCoverage always returns a
// non-nil map even when profile and functions are both empty.
func TestMapCoverage_ReturnsNonNilMap(t *testing.T) {
	result := coverage.MapCoverage(coverage.Profile{}, nil, nil)
	if result == nil {
		t.Error("MapCoverage returned nil map; want non-nil (even if empty)")
	}
}

// TestMapCoverage_NoFunctions will verify that when no functions are provided
// the returned map is empty.
func TestMapCoverage_NoFunctions(t *testing.T) {
	result := coverage.MapCoverage(coverage.Profile{Mode: "set"}, nil, nil)
	if len(result) != 0 {
		t.Errorf("MapCoverage with no functions = map of len %d; want 0", len(result))
	}
}

// TestBlock_FieldsExist is a compile-time guard that all expected fields on
// coverage.Block are present and accessible.
func TestBlock_FieldsExist(t *testing.T) {
	b := coverage.Block{
		File:      "pkg/foo.go",
		StartLine: 10,
		EndLine:   20,
		StartCol:  1,
		EndCol:    50,
		NumStmts:  5,
		Count:     1,
	}
	if b.File != "pkg/foo.go" {
		t.Errorf("Block.File = %q; want %q", b.File, "pkg/foo.go")
	}
}

// TestProfile_FieldsExist is a compile-time guard for coverage.Profile fields.
func TestProfile_FieldsExist(t *testing.T) {
	p := coverage.Profile{
		Mode:   "set",
		Blocks: []coverage.Block{},
	}
	if p.Mode != "set" {
		t.Errorf("Profile.Mode = %q; want \"set\"", p.Mode)
	}
}
