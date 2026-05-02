package main

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-crap4go/crap4go/internal/config"
	"github.com/go-crap4go/crap4go/internal/crap"
)

// fixturesDir returns the absolute path to testdata/pipeline/.
func fixturesDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile = cmd/crap/main_test.go; module root is two levels up
	moduleRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return filepath.Join(moduleRoot, "testdata", "pipeline")
}

// fixtureFile returns the absolute path to a file in testdata/pipeline/.
func fixtureFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(fixturesDir(t), name)
}

// ---------------------------------------------------------------------------
// filterEntries
// ---------------------------------------------------------------------------

// TestFilterEntries_ShowAll verifies that showAll=true returns all entries
// including low-risk ones.
func TestFilterEntries_ShowAll(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Low", Score: 1.0},     // low risk
		{Name: "Moderate", Score: 10}, // moderate risk
		{Name: "High", Score: 30},     // high risk
	}
	got := filterEntries(entries, true)
	if len(got) != 3 {
		t.Errorf("filterEntries(showAll=true) len=%d; want 3", len(got))
	}
}

// TestFilterEntries_HideLow verifies that showAll=false filters out low-risk
// entries (score < 5).
func TestFilterEntries_HideLow(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Low", Score: 1.0},
		{Name: "Moderate", Score: 10},
		{Name: "High", Score: 30},
	}
	got := filterEntries(entries, false)
	if len(got) != 2 {
		t.Errorf("filterEntries(showAll=false) len=%d; want 2 (low removed)", len(got))
	}
	for _, e := range got {
		if e.Name == "Low" {
			t.Errorf("filterEntries should have removed low-risk entry 'Low'")
		}
	}
}

// TestFilterEntries_AllLow verifies that when all entries are low-risk and
// showAll=false, the result is empty.
func TestFilterEntries_AllLow(t *testing.T) {
	entries := []crap.Entry{
		{Name: "A", Score: 1.0},
		{Name: "B", Score: 2.0},
	}
	got := filterEntries(entries, false)
	if len(got) != 0 {
		t.Errorf("filterEntries all-low got %d entries; want 0", len(got))
	}
}

// TestFilterEntries_Empty verifies that an empty input produces empty output.
func TestFilterEntries_Empty(t *testing.T) {
	got := filterEntries(nil, false)
	if len(got) != 0 {
		t.Errorf("filterEntries nil input len=%d; want 0", len(got))
	}
}

// ---------------------------------------------------------------------------
// countHighRisk
// ---------------------------------------------------------------------------

// TestCountHighRisk_None verifies zero when no entry meets the threshold.
func TestCountHighRisk_None(t *testing.T) {
	entries := []crap.Entry{{Score: 5}, {Score: 29}}
	got := countHighRisk(entries, 30)
	if got != 0 {
		t.Errorf("countHighRisk = %d; want 0", got)
	}
}

// TestCountHighRisk_Some verifies the correct count of entries at or above
// threshold.
func TestCountHighRisk_Some(t *testing.T) {
	entries := []crap.Entry{{Score: 30}, {Score: 50}, {Score: 10}}
	got := countHighRisk(entries, 30)
	if got != 2 {
		t.Errorf("countHighRisk = %d; want 2", got)
	}
}

// TestCountHighRisk_All verifies when all entries exceed threshold.
func TestCountHighRisk_All(t *testing.T) {
	entries := []crap.Entry{{Score: 100}, {Score: 200}}
	got := countHighRisk(entries, 30)
	if got != 2 {
		t.Errorf("countHighRisk = %d; want 2", got)
	}
}

// ---------------------------------------------------------------------------
// run — flag validation
// ---------------------------------------------------------------------------

// TestRun_MutuallyExclusiveFlags verifies that --run-tests and --no-run-tests
// together return an error.
func TestRun_MutuallyExclusiveFlags(t *testing.T) {
	cfg := config.Config{
		RunTests:   true,
		NoRunTests: true,
	}
	err := run(cfg)
	if err == nil {
		t.Error("run with both --run-tests and --no-run-tests returned nil error; want error")
	}
}

// TestRun_NoRunTestsWithoutProfile verifies that --no-run-tests without a
// coverage profile returns an error.
func TestRun_NoRunTestsWithoutProfile(t *testing.T) {
	cfg := config.Config{
		NoRunTests:   true,
		CoverProfile: "",
	}
	err := run(cfg)
	if err == nil {
		t.Error("run with --no-run-tests and no coverprofile returned nil error; want error")
	}
}

// ---------------------------------------------------------------------------
// buildRootCommand
// ---------------------------------------------------------------------------

// TestBuildRootCommand_FlagsExist verifies that all expected flags are
// registered on the root command.
func TestBuildRootCommand_FlagsExist(t *testing.T) {
	cmd := buildRootCommand()
	flags := []string{"coverprofile", "run-tests", "no-run-tests", "threshold", "config", "all"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag --%s not registered on root command", name)
		}
	}
}

// TestBuildRootCommand_DefaultThreshold verifies that the threshold flag
// defaults to 30.
func TestBuildRootCommand_DefaultThreshold(t *testing.T) {
	cmd := buildRootCommand()
	f := cmd.Flags().Lookup("threshold")
	if f == nil {
		t.Fatal("--threshold flag not found")
	}
	if f.DefValue != "30" {
		t.Errorf("--threshold default = %q; want \"30\"", f.DefValue)
	}
}

// ---------------------------------------------------------------------------
// run — happy path with testdata
// ---------------------------------------------------------------------------

// TestRun_HappyPath_WithCoverProfile exercises the full run() pipeline using
// testdata fixtures. Simple (CC=1) and Branch (CC=2) produce low CRAP scores,
// so the function should return nil (no threshold exceeded).
func TestRun_HappyPath_WithCoverProfile(t *testing.T) {
	cfg := config.Config{
		Dir:          fixturesDir(t),
		Paths:        []string{"foo"},
		CoverProfile: fixtureFile(t, "foo.out"),
		Threshold:    30,
	}
	err := run(cfg)
	if err != nil {
		t.Errorf("run returned error: %v; want nil (all entries below threshold)", err)
	}
}

// TestRun_ThresholdExceeded verifies that run() returns an error when a
// function's score meets or exceeds the threshold.
func TestRun_ThresholdExceeded(t *testing.T) {
	// Use threshold=2: Simple (CC=1, no cov) → CRAP=2, Branch (CC=2, no cov) → CRAP=6
	// Both exceed threshold=2, so run() should return an error.
	cfg := config.Config{
		Dir:        fixturesDir(t),
		Paths:      []string{"foo"},
		NoRunTests: true,
		Threshold:  2,
		// No CoverProfile → but NoRunTests requires one; use ShowAll to avoid
		// the no-run-tests validation by providing a CoverProfile.
		CoverProfile: fixtureFile(t, "foo.out"),
	}
	err := run(cfg)
	// Branch with partial coverage → CRAP score = 2²×(1−0.5)³+2 = 4×0.125+2 = 2.5
	// Simple with full coverage (1.0) → CRAP score = 1²×(1−1.0)³+1 = 0+1 = 1
	// With threshold=2: only Branch (2.5) exceeds threshold=2. Expect error.
	if err == nil {
		t.Error("run with low threshold returned nil error; want error")
	}
}
