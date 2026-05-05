package complexity_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-crap4go/crap4go/internal/complexity"
)

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

// mustParseFile parses a file from the testdata/complexity directory.
func mustParseFile(t *testing.T, name string) (*token.FileSet, *ast.File, string) {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "complexity")
	path := filepath.Join(dir, name)
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("mustParseFile(%q): %v", name, err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		t.Fatalf("mustParseFile(%q): %v", name, err)
	}
	return fset, f, path
}

// findFunc finds a Function by name in the slice; returns nil if not found.
func findFunc(fns []*complexity.Function, name string) *complexity.Function {
	for _, f := range fns {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// API contract (compile-time + nil guarantees)
// ---------------------------------------------------------------------------

// TestExtract_EmptyFile verifies that Extract returns no functions for an
// empty (package-only) source file without returning an error.
func TestExtract_EmptyFile(t *testing.T) {
	fset, file := mustParse(t, `package empty`)

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
	fset, file := mustParse(t, `package empty`)

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
	if f.StartLine != 1 || f.EndLine != 10 {
		t.Errorf("StartLine=%d EndLine=%d; want 1, 10", f.StartLine, f.EndLine)
	}
}

// ---------------------------------------------------------------------------
// Base CC = 1
// ---------------------------------------------------------------------------

func TestExtract_SimpleFunction_BaseCC(t *testing.T) {
	fset, file, path := mustParseFile(t, "simple.go")

	fns, err := complexity.Extract(fset, file, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := findFunc(fns, "Add")
	if f == nil {
		t.Fatalf("function Add not found; got: %v", funcNames(fns))
	}
	if f.CC != 1 {
		t.Errorf("Add CC = %d; want 1", f.CC)
	}
}

func TestExtract_SimpleFunction_LOC(t *testing.T) {
	fset, file, path := mustParseFile(t, "simple.go")

	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "Add")
	if f == nil {
		t.Fatal("function Add not found")
	}
	if f.LOC <= 0 {
		t.Errorf("Add LOC = %d; want > 0", f.LOC)
	}
}

func TestExtract_SimpleFunction_LineRange(t *testing.T) {
	fset, file, path := mustParseFile(t, "simple.go")

	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "Add")
	if f == nil {
		t.Fatal("function Add not found")
	}
	if f.StartLine <= 0 || f.EndLine <= 0 || f.EndLine < f.StartLine {
		t.Errorf("Add line range [%d, %d] is invalid", f.StartLine, f.EndLine)
	}
}

// ---------------------------------------------------------------------------
// If statements
// ---------------------------------------------------------------------------

func TestExtract_IfOnly_CC2(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, err := complexity.Extract(fset, file, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := findFunc(fns, "IfOnly")
	if f == nil {
		t.Fatalf("IfOnly not found; got: %v", funcNames(fns))
	}
	if f.CC != 2 {
		t.Errorf("IfOnly CC = %d; want 2", f.CC)
	}
}

func TestExtract_ElseIf_CC3(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "ElseIf")
	if f == nil {
		t.Fatalf("ElseIf not found")
	}
	if f.CC != 3 {
		t.Errorf("ElseIf CC = %d; want 3", f.CC)
	}
}

// ---------------------------------------------------------------------------
// For loops
// ---------------------------------------------------------------------------

func TestExtract_ForLoop_CC2(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "ForLoop")
	if f == nil {
		t.Fatal("ForLoop not found")
	}
	if f.CC != 2 {
		t.Errorf("ForLoop CC = %d; want 2", f.CC)
	}
}

func TestExtract_RangeLoop_CC2(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "RangeLoop")
	if f == nil {
		t.Fatal("RangeLoop not found")
	}
	if f.CC != 2 {
		t.Errorf("RangeLoop CC = %d; want 2", f.CC)
	}
}

// ---------------------------------------------------------------------------
// Switch statements
// ---------------------------------------------------------------------------

func TestExtract_SwitchStmt_CC4(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "SwitchStmt")
	if f == nil {
		t.Fatal("SwitchStmt not found")
	}
	// base(1) + 3 non-default cases = 4
	if f.CC != 4 {
		t.Errorf("SwitchStmt CC = %d; want 4", f.CC)
	}
}

func TestExtract_TypeSwitch_CC3(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "TypeSwitch")
	if f == nil {
		t.Fatal("TypeSwitch not found")
	}
	// base(1) + 2 non-default cases = 3
	if f.CC != 3 {
		t.Errorf("TypeSwitch CC = %d; want 3", f.CC)
	}
}

// ---------------------------------------------------------------------------
// Select statement
// ---------------------------------------------------------------------------

func TestExtract_SelectFunc_CC3(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "SelectFunc")
	if f == nil {
		t.Fatal("SelectFunc not found")
	}
	// base(1) + 2 non-default comms = 3
	if f.CC != 3 {
		t.Errorf("SelectFunc CC = %d; want 3", f.CC)
	}
}

// ---------------------------------------------------------------------------
// Goto
// ---------------------------------------------------------------------------

func TestExtract_WithGoto_CC3(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "WithGoto")
	if f == nil {
		t.Fatal("WithGoto not found")
	}
	// base(1) + if(1) + goto(1) = 3
	if f.CC != 3 {
		t.Errorf("WithGoto CC = %d; want 3", f.CC)
	}
}

// ---------------------------------------------------------------------------
// Logical operators
// ---------------------------------------------------------------------------

func TestExtract_AndOr_CC3(t *testing.T) {
	fset, file, path := mustParseFile(t, "logical.go")
	fns, err := complexity.Extract(fset, file, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := findFunc(fns, "AndOr")
	if f == nil {
		t.Fatalf("AndOr not found; got: %v", funcNames(fns))
	}
	// base(1) + &&(1) + ||(1) = 3
	if f.CC != 3 {
		t.Errorf("AndOr CC = %d; want 3", f.CC)
	}
}

func TestExtract_NestedLogic_CC4(t *testing.T) {
	fset, file, path := mustParseFile(t, "logical.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "NestedLogic")
	if f == nil {
		t.Fatal("NestedLogic not found")
	}
	// base(1) + &&(1) + ||(1) + &&(1) = 4
	if f.CC != 4 {
		t.Errorf("NestedLogic CC = %d; want 4", f.CC)
	}
}

func TestExtract_InIf_CC3(t *testing.T) {
	fset, file, path := mustParseFile(t, "logical.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "InIf")
	if f == nil {
		t.Fatal("InIf not found")
	}
	// base(1) + if(1) + &&(1) = 3
	if f.CC != 3 {
		t.Errorf("InIf CC = %d; want 3", f.CC)
	}
}

// TestExtract_BoolAssign_CC2 pins the condition-agnostic &&/|| counting
// behaviour: the tool counts logical operators regardless of whether they
// appear in a branch condition or a plain assignment/return. This is a
// documented design decision — not a bug.
func TestExtract_BoolAssign_CC2(t *testing.T) {
	fset, file, path := mustParseFile(t, "logical.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "BoolAssign")
	if f == nil {
		t.Fatal("BoolAssign not found")
	}
	// base(1) + &&(1) = 2  — && is in an assignment, not a branch condition
	if f.CC != 2 {
		t.Errorf("BoolAssign CC = %d; want 2 (condition-agnostic && counting)", f.CC)
	}
}

// TestExtract_BoolReturn_CC2 pins condition-agnostic || counting in a return.
func TestExtract_BoolReturn_CC2(t *testing.T) {
	fset, file, path := mustParseFile(t, "logical.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "BoolReturn")
	if f == nil {
		t.Fatal("BoolReturn not found")
	}
	// base(1) + ||(1) = 2  — || is in a return, not a branch condition
	if f.CC != 2 {
		t.Errorf("BoolReturn CC = %d; want 2 (condition-agnostic || counting)", f.CC)
	}
}

// ---------------------------------------------------------------------------
// Methods
// ---------------------------------------------------------------------------

func TestExtract_ValueReceiverMethod(t *testing.T) {
	fset, file, path := mustParseFile(t, "methods.go")
	fns, err := complexity.Extract(fset, file, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := findFunc(fns, "Greeter.Greet")
	if f == nil {
		t.Fatalf("Greeter.Greet not found; got: %v", funcNames(fns))
	}
	if f.CC != 1 {
		t.Errorf("Greeter.Greet CC = %d; want 1", f.CC)
	}
}

func TestExtract_PointerReceiverNormalised(t *testing.T) {
	// Pointer receiver *Greeter should be displayed as "Greeter.SetPrefix"
	fset, file, path := mustParseFile(t, "methods.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "Greeter.SetPrefix")
	if f == nil {
		t.Fatalf("Greeter.SetPrefix not found (pointer receiver not normalised); got: %v", funcNames(fns))
	}
}

func TestExtract_MethodWithIf_CC2(t *testing.T) {
	fset, file, path := mustParseFile(t, "methods.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "Greeter.ConditionalGreet")
	if f == nil {
		t.Fatalf("Greeter.ConditionalGreet not found; got: %v", funcNames(fns))
	}
	if f.CC != 2 {
		t.Errorf("Greeter.ConditionalGreet CC = %d; want 2", f.CC)
	}
}

func TestExtract_StandaloneFunc(t *testing.T) {
	fset, file, path := mustParseFile(t, "methods.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "StandaloneFunc")
	if f == nil {
		t.Fatalf("StandaloneFunc not found; got: %v", funcNames(fns))
	}
	if f.CC != 1 {
		t.Errorf("StandaloneFunc CC = %d; want 1", f.CC)
	}
}

// ---------------------------------------------------------------------------
// Closures
// ---------------------------------------------------------------------------

func TestExtract_Closure_SeparateEntry(t *testing.T) {
	fset, file, path := mustParseFile(t, "closures.go")
	fns, err := complexity.Extract(fset, file, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outer := findFunc(fns, "WithClosure")
	if outer == nil {
		t.Fatalf("WithClosure not found; got: %v", funcNames(fns))
	}
	// Should have at least one anonymous entry for the closure.
	var anonCount int
	for _, fn := range fns {
		if len(fn.Name) > 0 && fn.Name[0] == '<' {
			anonCount++
		}
	}
	if anonCount == 0 {
		t.Errorf("no anonymous (closure) entries found; got: %v", funcNames(fns))
	}
}

func TestExtract_Closure_DoesNotContaminateOuter(t *testing.T) {
	// WithClosure: outer body has one if statement → CC = 2.
	// The closure has its own if statement but must NOT add to outer CC.
	fset, file, path := mustParseFile(t, "closures.go")
	fns, _ := complexity.Extract(fset, file, path)
	outer := findFunc(fns, "WithClosure")
	if outer == nil {
		t.Fatal("WithClosure not found")
	}
	if outer.CC != 2 {
		t.Errorf("WithClosure CC = %d; want 2 (closure branches must not contaminate outer)", outer.CC)
	}
}

func TestExtract_MultiClosure_OuterCC1(t *testing.T) {
	// MultiClosure outer body has no branches → CC = 1.
	fset, file, path := mustParseFile(t, "closures.go")
	fns, _ := complexity.Extract(fset, file, path)
	outer := findFunc(fns, "MultiClosure")
	if outer == nil {
		t.Fatal("MultiClosure not found")
	}
	if outer.CC != 1 {
		t.Errorf("MultiClosure CC = %d; want 1", outer.CC)
	}
}

func TestExtract_MultiClosure_TwoAnonEntries(t *testing.T) {
	// MultiClosure has two closures; both must appear as separate entries.
	fset, file, path := mustParseFile(t, "closures.go")
	fns, _ := complexity.Extract(fset, file, path)

	var anonInMulti []*complexity.Function
	multiOuter := findFunc(fns, "MultiClosure")
	if multiOuter == nil {
		t.Fatal("MultiClosure not found")
	}
	for _, fn := range fns {
		if len(fn.Name) > 0 && fn.Name[0] == '<' &&
			fn.StartLine >= multiOuter.StartLine &&
			fn.EndLine <= multiOuter.EndLine {
			anonInMulti = append(anonInMulti, fn)
		}
	}
	if len(anonInMulti) != 2 {
		t.Errorf("expected 2 anonymous entries inside MultiClosure; got %d", len(anonInMulti))
	}
}

// ---------------------------------------------------------------------------
// MixedBranches (integration)
// ---------------------------------------------------------------------------

func TestExtract_MixedBranches_CC4(t *testing.T) {
	fset, file, path := mustParseFile(t, "branches.go")
	fns, _ := complexity.Extract(fset, file, path)
	f := findFunc(fns, "MixedBranches")
	if f == nil {
		t.Fatalf("MixedBranches not found; got: %v", funcNames(fns))
	}
	// base(1) + if(1) + for(1) + range(1) = 4
	if f.CC != 4 {
		t.Errorf("MixedBranches CC = %d; want 4", f.CC)
	}
}

// ---------------------------------------------------------------------------
// File field
// ---------------------------------------------------------------------------

func TestExtract_FileFieldSet(t *testing.T) {
	fset, file, path := mustParseFile(t, "simple.go")
	fns, _ := complexity.Extract(fset, file, path)
	for _, fn := range fns {
		if fn.File != path {
			t.Errorf("Function.File = %q; want %q", fn.File, path)
		}
	}
}

// ---------------------------------------------------------------------------
// init() deduplication
// ---------------------------------------------------------------------------

// TestExtract_MultipleInit_UniqueNames verifies that when a file contains
// multiple func init() declarations they are given unique names
// (init, init#2, init#3) so that report output is unambiguous and
// MapCoverage does not overwrite earlier entries.
func TestExtract_MultipleInit_UniqueNames(t *testing.T) {
	fset, file, path := mustParseFile(t, "initfuncs.go")
	fns, err := complexity.Extract(fset, file, path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	names := funcNames(fns)

	// Expect exactly three entries, all from init() declarations.
	var initFns []*complexity.Function
	for _, f := range fns {
		if f.Name == "init" || len(f.Name) > 4 && f.Name[:4] == "init" {
			initFns = append(initFns, f)
		}
	}
	if len(initFns) != 3 {
		t.Fatalf("expected 3 init entries; got %d: %v", len(initFns), names)
	}

	// All three names must be distinct.
	seen := make(map[string]bool, 3)
	for _, f := range initFns {
		if seen[f.Name] {
			t.Errorf("duplicate init name %q; all names: %v", f.Name, names)
		}
		seen[f.Name] = true
	}
}

// ---------------------------------------------------------------------------
// helper
// ---------------------------------------------------------------------------

func funcNames(fns []*complexity.Function) []string {
	names := make([]string, len(fns))
	for i, f := range fns {
		names[i] = f.Name
	}
	return names
}
