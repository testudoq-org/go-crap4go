package pipeline_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/go-crap4go/crap4go/internal/config"
	"github.com/go-crap4go/crap4go/internal/crap"
	"github.com/go-crap4go/crap4go/internal/pipeline"
)

// ---------------------------------------------------------------------------
// path helpers
// ---------------------------------------------------------------------------

// fixturesDir returns the absolute path to testdata/pipeline/ — the root of
// all pipeline test fixtures. Walking from here avoids the "skip testdata"
// guard in collectFiles because the directory itself is named "pipeline".
func fixturesDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	moduleRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return filepath.Join(moduleRoot, "testdata", "pipeline")
}

// fixtureFile returns the absolute path to a named file under testdata/pipeline/.
func fixtureFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(fixturesDir(t), name)
}

// ---------------------------------------------------------------------------
// TestAnalyse_FindsFunction
// ---------------------------------------------------------------------------

// TestAnalyse_FindsFunction verifies that Analyse returns an entry for the
// "Simple" function in the testdata/pipeline/foo/foo.go fixture.
func TestAnalyse_FindsFunction(t *testing.T) {
	cfg := config.Config{
		Dir:   fixturesDir(t),
		Paths: []string{"foo"},
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("Analyse returned 0 entries; want at least 1")
	}
	var found bool
	for _, e := range entries {
		if e.Name == "Simple" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("entry for 'Simple' not found; got: %v", entryNames(entries))
	}
}

// TestAnalyse_FindsBranchFunction verifies that Analyse returns an entry for
// the "Branch" function (CC=2).
func TestAnalyse_FindsBranchFunction(t *testing.T) {
	cfg := config.Config{
		Dir:   fixturesDir(t),
		Paths: []string{"foo"},
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	var e *crap.Entry
	for i := range entries {
		if entries[i].Name == "Branch" {
			e = &entries[i]
			break
		}
	}
	if e == nil {
		t.Fatalf("entry for 'Branch' not found; got: %v", entryNames(entries))
	}
	if e.CC != 2 {
		t.Errorf("Branch.CC = %d; want 2", e.CC)
	}
}

// ---------------------------------------------------------------------------
// TestAnalyse_PathFilter
// ---------------------------------------------------------------------------

// TestAnalyse_PathFilter verifies that a path fragment filter restricts
// analysis to matching files only.
func TestAnalyse_PathFilter(t *testing.T) {
	cfg := config.Config{
		Dir:   fixturesDir(t),
		Paths: []string{"foo"},
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	for _, e := range entries {
		if !strings.Contains(filepath.ToSlash(e.File), "foo") {
			t.Errorf("entry %q in file %q should not be included by 'foo' path filter",
				e.Name, e.File)
		}
	}
}

// ---------------------------------------------------------------------------
// TestAnalyse_SkipsTestFiles
// ---------------------------------------------------------------------------

// TestAnalyse_SkipsTestFiles verifies that _test.go files are excluded from
// the analysis.
func TestAnalyse_SkipsTestFiles(t *testing.T) {
	cfg := config.Config{
		Dir:   fixturesDir(t),
		Paths: []string{"foo"},
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	// foo_test.go defines TestSimple; it must not appear in results.
	for _, e := range entries {
		if e.Name == "TestSimple" {
			t.Errorf("TestSimple should be excluded (_test.go); found in entries")
		}
	}
}

// ---------------------------------------------------------------------------
// TestAnalyse_NoCoverProfile_SentinelCoverage
// ---------------------------------------------------------------------------

// TestAnalyse_NoCoverProfile_SentinelCoverage verifies that without a
// coverage profile every entry has Coverage == -1 (sentinel = no data).
func TestAnalyse_NoCoverProfile_SentinelCoverage(t *testing.T) {
	cfg := config.Config{
		Dir:   fixturesDir(t),
		Paths: []string{"foo"},
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	for _, e := range entries {
		if e.Coverage != -1 {
			t.Errorf("entry %q.Coverage = %f; want -1 (no profile)", e.Name, e.Coverage)
		}
	}
}

// ---------------------------------------------------------------------------
// TestAnalyse_WithCoverProfile
// ---------------------------------------------------------------------------

// TestAnalyse_WithCoverProfile verifies that when a coverage profile is
// supplied, entries have their Coverage populated (0–1 range).
func TestAnalyse_WithCoverProfile(t *testing.T) {
	cfg := config.Config{
		Dir:          fixturesDir(t),
		Paths:        []string{"foo"},
		CoverProfile: fixtureFile(t, "foo.out"),
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	for _, e := range entries {
		if e.Coverage == -1 {
			t.Errorf("entry %q.Coverage = -1; profile should provide data", e.Name)
		}
		if e.Coverage < 0 || e.Coverage > 1 {
			t.Errorf("entry %q.Coverage = %f; want 0 <= cov <= 1", e.Name, e.Coverage)
		}
	}
}

// ---------------------------------------------------------------------------
// TestAnalyse_EntryFields
// ---------------------------------------------------------------------------

// TestAnalyse_EntryFields verifies that every entry in the result has all
// expected fields populated.
func TestAnalyse_EntryFields(t *testing.T) {
	cfg := config.Config{
		Dir:   fixturesDir(t),
		Paths: []string{"foo"},
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Error("entry.Name is empty")
		}
		if e.File == "" {
			t.Error("entry.File is empty")
		}
		if e.CC < 1 {
			t.Errorf("entry %q.CC = %d; want >= 1", e.Name, e.CC)
		}
		if e.LOC <= 0 {
			t.Errorf("entry %q.LOC = %d; want > 0", e.Name, e.LOC)
		}
		if e.Score <= 0 {
			t.Errorf("entry %q.Score = %f; want > 0", e.Name, e.Score)
		}
	}
}

// ---------------------------------------------------------------------------
// TestAnalyse_ComputesCRAPScore
// ---------------------------------------------------------------------------

// TestAnalyse_ComputesCRAPScore verifies that the CRAP score equals
// CC²×(1−coverage)³+CC. With coverage=-1 (sentinel → 0%): CRAP = CC(CC+1).
func TestAnalyse_ComputesCRAPScore(t *testing.T) {
	cfg := config.Config{
		Dir:   fixturesDir(t),
		Paths: []string{"foo"},
	}
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	for _, e := range entries {
		want := float64(e.CC*e.CC) + float64(e.CC)
		if absf(e.Score-want) > 1e-9 {
			t.Errorf("entry %q: Score = %f; want CC(%d)²+CC = %f",
				e.Name, e.Score, e.CC, want)
		}
	}
}

// ---------------------------------------------------------------------------
// TestAnalyse_BadCoverProfile_Error
// ---------------------------------------------------------------------------

// TestAnalyse_BadCoverProfile_Error verifies that a non-existent coverage
// profile path causes Analyse to return an error.
func TestAnalyse_BadCoverProfile_Error(t *testing.T) {
	cfg := config.Config{
		Dir:          fixturesDir(t),
		Paths:        []string{"foo"},
		CoverProfile: "/does/not/exist.out",
	}
	_, err := pipeline.Analyse(cfg)
	if err == nil {
		t.Error("Analyse with bad CoverProfile returned nil error; want error")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func entryNames(entries []crap.Entry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	return names
}

func absf(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
