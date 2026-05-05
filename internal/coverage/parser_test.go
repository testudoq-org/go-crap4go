package coverage_test

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-crap4go/crap4go/internal/complexity"
	"github.com/go-crap4go/crap4go/internal/coverage"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "testdata", "coverage", name)
}

// ---------------------------------------------------------------------------
// LoadProfile
// ---------------------------------------------------------------------------

// TestLoadProfile_NonExistentFile verifies that LoadProfile returns an error
// for a file that does not exist.
func TestLoadProfile_NonExistentFile(t *testing.T) {
	_, err := coverage.LoadProfile("testdata/nonexistent.out")
	if err == nil {
		t.Error("LoadProfile returned nil error for non-existent file; want error")
	}
}

// TestLoadProfile_SimpleSet parses a "set" mode coverage profile and checks
// the Mode field and block count.
func TestLoadProfile_SimpleSet(t *testing.T) {
	p, err := coverage.LoadProfile(testdataPath(t, "simple.out"))
	if err != nil {
		t.Fatalf("LoadProfile simple.out: %v", err)
	}
	if p.Mode != "set" {
		t.Errorf("Mode = %q; want \"set\"", p.Mode)
	}
	if len(p.Blocks) != 7 {
		t.Errorf("len(Blocks) = %d; want 7", len(p.Blocks))
	}
}

// TestLoadProfile_CountMode parses a "count" mode profile.
func TestLoadProfile_CountMode(t *testing.T) {
	p, err := coverage.LoadProfile(testdataPath(t, "count.out"))
	if err != nil {
		t.Fatalf("LoadProfile count.out: %v", err)
	}
	if p.Mode != "count" {
		t.Errorf("Mode = %q; want \"count\"", p.Mode)
	}
	if len(p.Blocks) != 4 {
		t.Errorf("len(Blocks) = %d; want 4", len(p.Blocks))
	}
}

// TestLoadProfile_AtomicMode parses an "atomic" mode profile.
func TestLoadProfile_AtomicMode(t *testing.T) {
	p, err := coverage.LoadProfile(testdataPath(t, "atomic.out"))
	if err != nil {
		t.Fatalf("LoadProfile atomic.out: %v", err)
	}
	if p.Mode != "atomic" {
		t.Errorf("Mode = %q; want \"atomic\"", p.Mode)
	}
}

// TestLoadProfile_InvalidFile returns an error for a malformed profile.
func TestLoadProfile_InvalidFile(t *testing.T) {
	_, err := coverage.LoadProfile(testdataPath(t, "invalid.out"))
	if err == nil {
		t.Error("LoadProfile invalid.out returned nil error; want error")
	}
}

// TestLoadProfile_BlockFields verifies that a parsed Block has the correct
// numeric fields for the first line of simple.out:
//   github.com/go-crap4go/crap4go/internal/crap/crap.go:55.39,59.16 3 1
func TestLoadProfile_BlockFields(t *testing.T) {
	p, err := coverage.LoadProfile(testdataPath(t, "simple.out"))
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}
	if len(p.Blocks) == 0 {
		t.Fatal("no blocks parsed")
	}
	b := p.Blocks[0]
	if b.StartLine != 55 {
		t.Errorf("Block[0].StartLine = %d; want 55", b.StartLine)
	}
	if b.EndLine != 59 {
		t.Errorf("Block[0].EndLine = %d; want 59", b.EndLine)
	}
	if b.StartCol != 39 {
		t.Errorf("Block[0].StartCol = %d; want 39", b.StartCol)
	}
	if b.EndCol != 16 {
		t.Errorf("Block[0].EndCol = %d; want 16", b.EndCol)
	}
	if b.NumStmts != 3 {
		t.Errorf("Block[0].NumStmts = %d; want 3", b.NumStmts)
	}
	if b.Count != 1 {
		t.Errorf("Block[0].Count = %d; want 1", b.Count)
	}
}

// TestLoadProfile_UncoveredBlock verifies that a block with Count=0 is
// correctly parsed with Count = 0 (not dropped).
func TestLoadProfile_UncoveredBlock(t *testing.T) {
	p, err := coverage.LoadProfile(testdataPath(t, "simple.out"))
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}
	// simple.out line 2: lines 59-61, Count=0
	b := p.Blocks[1]
	if b.Count != 0 {
		t.Errorf("Block[1].Count = %d; want 0", b.Count)
	}
}

// ---------------------------------------------------------------------------
// Struct field guards
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// MapCoverage
// ---------------------------------------------------------------------------

// TestMapCoverage_ReturnsNonNilMap verifies that MapCoverage always returns a
// non-nil map even when profile and functions are both empty.
func TestMapCoverage_ReturnsNonNilMap(t *testing.T) {
	result := coverage.MapCoverage(coverage.Profile{}, nil, nil)
	if result == nil {
		t.Error("MapCoverage returned nil map; want non-nil (even if empty)")
	}
}

// TestMapCoverage_NoFunctions verifies that an empty function list produces an
// empty map.
func TestMapCoverage_NoFunctions(t *testing.T) {
	result := coverage.MapCoverage(coverage.Profile{Mode: "set"}, nil, nil)
	if len(result) != 0 {
		t.Errorf("MapCoverage with no functions returned map of len %d; want 0", len(result))
	}
}

// TestMapCoverage_NoMatchingBlocks verifies that a function with no coverage
// blocks in its line range is mapped to -1.
func TestMapCoverage_NoMatchingBlocks(t *testing.T) {
	profile := coverage.Profile{
		Mode: "set",
		Blocks: []coverage.Block{
			{File: "other/file.go", StartLine: 1, EndLine: 5, NumStmts: 3, Count: 1},
		},
	}
	fns := []*complexity.Function{
		{Name: "MyFunc", File: "mypackage/myfile.go", StartLine: 10, EndLine: 20},
	}
	result := coverage.MapCoverage(profile, fns, nil)
	key := "mypackage/myfile.go:MyFunc"
	got, ok := result[key]
	if !ok {
		t.Fatalf("key %q not found in result map; got keys: %v", key, mapKeys(result))
	}
	if got != -1 {
		t.Errorf("coverage[%q] = %f; want -1 (no data)", key, got)
	}
}

// TestMapCoverage_FullyCovered verifies that when all blocks in a function's
// range have Count > 0, the coverage fraction is 1.0.
func TestMapCoverage_FullyCovered(t *testing.T) {
	file := "mypkg/myfile.go"
	profile := coverage.Profile{
		Mode: "set",
		Blocks: []coverage.Block{
			{File: file, StartLine: 10, EndLine: 15, NumStmts: 3, Count: 1},
			{File: file, StartLine: 15, EndLine: 20, NumStmts: 2, Count: 1},
		},
	}
	fns := []*complexity.Function{
		{Name: "Covered", File: file, StartLine: 10, EndLine: 20},
	}
	result := coverage.MapCoverage(profile, fns, nil)
	key := file + ":Covered"
	got, ok := result[key]
	if !ok {
		t.Fatalf("key %q not found", key)
	}
	if got != 1.0 {
		t.Errorf("coverage[%q] = %f; want 1.0", key, got)
	}
}

// TestMapCoverage_PartiallyCovered verifies that partial coverage is computed
// correctly as coveredStmts / totalStmts.
func TestMapCoverage_PartiallyCovered(t *testing.T) {
	file := "mypkg/partial.go"
	profile := coverage.Profile{
		Mode: "set",
		Blocks: []coverage.Block{
			{File: file, StartLine: 5, EndLine: 10, NumStmts: 4, Count: 1}, // covered
			{File: file, StartLine: 10, EndLine: 15, NumStmts: 1, Count: 0}, // not covered
		},
	}
	fns := []*complexity.Function{
		{Name: "Partial", File: file, StartLine: 5, EndLine: 15},
	}
	result := coverage.MapCoverage(profile, fns, nil)
	key := file + ":Partial"
	got, ok := result[key]
	if !ok {
		t.Fatalf("key %q not found", key)
	}
	// 4 covered out of 5 total = 0.8
	want := 0.8
	if abs(got-want) > 1e-9 {
		t.Errorf("coverage[%q] = %f; want %f", key, got, want)
	}
}

// TestMapCoverage_ZeroCoverage verifies that a function whose blocks all have
// Count == 0 is mapped to 0.0 (not -1).
func TestMapCoverage_ZeroCoverage(t *testing.T) {
	file := "mypkg/uncovered.go"
	profile := coverage.Profile{
		Mode: "set",
		Blocks: []coverage.Block{
			{File: file, StartLine: 1, EndLine: 10, NumStmts: 5, Count: 0},
		},
	}
	fns := []*complexity.Function{
		{Name: "NeverCalled", File: file, StartLine: 1, EndLine: 10},
	}
	result := coverage.MapCoverage(profile, fns, nil)
	key := file + ":NeverCalled"
	got, ok := result[key]
	if !ok {
		t.Fatalf("key %q not found", key)
	}
	if got != 0.0 {
		t.Errorf("coverage[%q] = %f; want 0.0", key, got)
	}
}

// TestMapCoverage_MultipleFunctions verifies independent coverage per function.
func TestMapCoverage_MultipleFunctions(t *testing.T) {
	file := "mypkg/multi.go"
	profile := coverage.Profile{
		Mode: "set",
		Blocks: []coverage.Block{
			{File: file, StartLine: 1, EndLine: 5, NumStmts: 2, Count: 1},
			{File: file, StartLine: 10, EndLine: 15, NumStmts: 3, Count: 0},
		},
	}
	fns := []*complexity.Function{
		{Name: "FuncA", File: file, StartLine: 1, EndLine: 5},
		{Name: "FuncB", File: file, StartLine: 10, EndLine: 15},
	}
	result := coverage.MapCoverage(profile, fns, nil)

	keyA := file + ":FuncA"
	keyB := file + ":FuncB"

	if result[keyA] != 1.0 {
		t.Errorf("FuncA coverage = %f; want 1.0", result[keyA])
	}
	if result[keyB] != 0.0 {
		t.Errorf("FuncB coverage = %f; want 0.0", result[keyB])
	}
}

// TestMapCoverage_ProfileFileMatchesSuffix verifies that block file paths in
// coverage profiles (which use the full module path) are matched against
// Function.File using a suffix/contains check.
func TestMapCoverage_ProfileFileMatchesSuffix(t *testing.T) {
	// Coverage profiles record full module paths; function files are typically
	// workspace-relative. MapCoverage must handle partial path matching.
	blockFile := "github.com/go-crap4go/crap4go/internal/crap/crap.go"
	funcFile := "internal/crap/crap.go"

	profile := coverage.Profile{
		Mode: "set",
		Blocks: []coverage.Block{
			{File: blockFile, StartLine: 1, EndLine: 10, NumStmts: 3, Count: 1},
		},
	}
	fns := []*complexity.Function{
		{Name: "Score", File: funcFile, StartLine: 1, EndLine: 10},
	}

	// We need a FileSet for line-based matching; use a minimal one.
	fset := token.NewFileSet()
	fset.AddFile(funcFile, -1, 1000)

	result := coverage.MapCoverage(profile, fns, fset)
	key := funcFile + ":Score"
	got, ok := result[key]
	if !ok {
		t.Fatalf("key %q not found; got keys: %v", key, mapKeys(result))
	}
	if got != 1.0 {
		t.Errorf("coverage[%q] = %f; want 1.0", key, got)
	}
}

// ---------------------------------------------------------------------------
// Integration: LoadProfile + MapCoverage pipeline
// ---------------------------------------------------------------------------

// TestCoveragePipeline_SetMode runs a full load+map cycle on simple.out
// against a synthetic function list that spans the file's line ranges.
func TestCoveragePipeline_SetMode(t *testing.T) {
	p, err := coverage.LoadProfile(testdataPath(t, "simple.out"))
	if err != nil {
		t.Fatalf("LoadProfile: %v", err)
	}

	// Build a minimal function that spans lines 55-90 (covers all test blocks).
	src := "package crap\nfunc Score() {}\n"
	fset := token.NewFileSet()
	f, parseErr := parser.ParseFile(fset, "internal/crap/crap.go", src, 0)
	if parseErr != nil {
		t.Fatalf("parse: %v", parseErr)
	}
	_ = f

	fns := []*complexity.Function{
		{
			Name:      "Score",
			File:      "internal/crap/crap.go",
			StartLine: 50,
			EndLine:   90,
		},
	}

	result := coverage.MapCoverage(p, fns, fset)
	key := "internal/crap/crap.go:Score"
	got, ok := result[key]
	if !ok {
		t.Fatalf("key %q not found; keys: %v", key, mapKeys(result))
	}
	// simple.out has 5 covered out of 7 total statements → ~0.714
	if got < 0 || got > 1 {
		t.Errorf("coverage[%q] = %f; want 0 <= cov <= 1", key, got)
	}
}

// ---------------------------------------------------------------------------
// Error path tests (boost coverage above 85%)
// ---------------------------------------------------------------------------

// TestLoadProfile_BadMode verifies that an unrecognised mode string returns
// an error.
func TestLoadProfile_BadMode(t *testing.T) {
	_, err := coverage.LoadProfile(testdataPath(t, "badmode.out"))
	if err == nil {
		t.Error("LoadProfile badmode.out returned nil error; want error for unknown mode")
	}
}

// TestLoadProfile_BadBlockLine verifies that a malformed block line returns
// an error.
func TestLoadProfile_BadBlockLine(t *testing.T) {
	_, err := coverage.LoadProfile(testdataPath(t, "badblock.out"))
	if err == nil {
		t.Error("LoadProfile badblock.out returned nil error; want error for bad block")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mapKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

