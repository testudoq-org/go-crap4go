package complexity_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/go-crap4go/crap4go/internal/complexity"
)

// TestExtract_EmptyFile verifies that Extract returns no functions for an
// empty (package-only) source file without returning an error.
func TestExtract_EmptyFile(t *testing.T) {
	src := `package empty`
	fset, file := mustParse(t, src)

	fns, err := complexity.Extract(fset, file, "empty.go")
	if err != nil {
		t.Fatalf("Extract returned unexpected error: %v", err)
	}
	if len(fns) != 0 {
		t.Errorf("Extract returned %d functions; want 0", len(fns))
	}
}

// TestExtract_ReturnsNonNilSlice verifies that Extract never returns a nil
// slice, even for source files with no functions.
func TestExtract_ReturnsNonNilSlice(t *testing.T) {
	src := `package empty`
	fset, file := mustParse(t, src)

	fns, _ := complexity.Extract(fset, file, "empty.go")
	if fns == nil {
		t.Error("Extract returned nil slice; want non-nil (even if empty)")
	}
}

// TestFunction_FieldsExist is a compile-time guard that all expected fields
// on complexity.Function are present and accessible.
func TestFunction_FieldsExist(t *testing.T) {
	f := complexity.Function{
		Name:      "Foo",
		File:      "foo.go",
		StartLine: 1,
		EndLine:   10,
		CC:        1,
		LOC:       5,
	}
	if f.Name != "Foo" {
		t.Errorf("Name = %q; want %q", f.Name, "Foo")
	}
}

// ---------------------------------------------------------------------------
// TODO(prompt-2): The following tests are stubs that will be expanded once
// the full CC extraction logic is implemented.
// ---------------------------------------------------------------------------

// TestExtract_SingleFunction_BaseCC will verify that a function with no
// branching constructs receives a CC of exactly 1.
func TestExtract_SingleFunction_BaseCC(t *testing.T) {
	t.Skip("TODO(prompt-2): implement CC extraction")
}

// TestExtract_IfStatement will verify that an if statement adds +1 to CC.
func TestExtract_IfStatement(t *testing.T) {
	t.Skip("TODO(prompt-2): implement CC extraction")
}

// TestExtract_ForLoop will verify that a for loop adds +1 to CC.
func TestExtract_ForLoop(t *testing.T) {
	t.Skip("TODO(prompt-2): implement CC extraction")
}

// TestExtract_LogicalOperators will verify that && and || each add +1 to CC.
func TestExtract_LogicalOperators(t *testing.T) {
	t.Skip("TODO(prompt-2): implement CC extraction")
}

// TestExtract_Method will verify that methods are named "ReceiverType.MethodName".
func TestExtract_Method(t *testing.T) {
	t.Skip("TODO(prompt-2): implement CC extraction")
}

// TestExtract_Closure will verify that closures are reported as separate entries.
func TestExtract_Closure(t *testing.T) {
	t.Skip("TODO(prompt-2): implement CC extraction")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mustParse(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("mustParse: %v", err)
	}
	return fset, f
}
