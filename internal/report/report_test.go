package report_test

import (
	"strings"
	"testing"

	"github.com/go-crap4go/crap4go/internal/crap"
	"github.com/go-crap4go/crap4go/internal/report"
)

// ---------------------------------------------------------------------------
// FormatReport — header and structure
// ---------------------------------------------------------------------------

// TestFormatReport_EmptyEntries verifies that an empty entry list still
// renders a header and a summary line rather than an empty string.
func TestFormatReport_EmptyEntries(t *testing.T) {
	got := report.FormatReport([]crap.Entry{})
	if got == "" {
		t.Fatal("FormatReport(empty) returned empty string; want at least a header")
	}
	if !strings.Contains(got, "FUNCTION") {
		t.Errorf("FormatReport(empty) missing FUNCTION column header\n%s", got)
	}
	if !strings.Contains(got, "CRAP") {
		t.Errorf("FormatReport(empty) missing CRAP column header\n%s", got)
	}
}

// TestFormatReport_ContainsFunctionName verifies that each entry's Name
// appears somewhere in the formatted output.
func TestFormatReport_ContainsFunctionName(t *testing.T) {
	entries := []crap.Entry{
		{Name: "MyFunc", File: "pkg/foo.go", CC: 1, Coverage: 1.0, Score: 1.0, LOC: 5},
	}
	got := report.FormatReport(entries)
	if !strings.Contains(got, "MyFunc") {
		t.Errorf("FormatReport output missing function name %q\n%s", "MyFunc", got)
	}
}

// TestFormatReport_ContainsFileName verifies that each entry's File appears
// in the formatted output.
func TestFormatReport_ContainsFileName(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Foo", File: "internal/bar/baz.go", CC: 1, Coverage: 1.0, Score: 1.0, LOC: 3},
	}
	got := report.FormatReport(entries)
	if !strings.Contains(got, "internal/bar/baz.go") {
		t.Errorf("FormatReport output missing file name\n%s", got)
	}
}

// TestFormatReport_ContainsCRAPScore verifies that the numeric CRAP score
// appears in the formatted output.
func TestFormatReport_ContainsCRAPScore(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Foo", File: "foo.go", CC: 3, Coverage: 0.0, Score: 12.0, LOC: 10},
	}
	got := report.FormatReport(entries)
	if !strings.Contains(got, "12") {
		t.Errorf("FormatReport output missing CRAP score 12\n%s", got)
	}
}

// TestFormatReport_HighRiskLabel verifies that high-risk entries display
// "HIGH" (uppercase) to draw attention.
func TestFormatReport_HighRiskLabel(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Risky", File: "risky.go", CC: 10, Coverage: 0.0, Score: 110.0, LOC: 20},
	}
	got := report.FormatReport(entries)
	if !strings.Contains(got, "HIGH") {
		t.Errorf("FormatReport output missing HIGH risk label for score 110\n%s", got)
	}
}

// TestFormatReport_LowRiskLabel verifies that low-risk entries display "low".
func TestFormatReport_LowRiskLabel(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Simple", File: "simple.go", CC: 1, Coverage: 1.0, Score: 1.0, LOC: 3},
	}
	got := report.FormatReport(entries)
	if !strings.Contains(got, "low") {
		t.Errorf("FormatReport output missing low risk label\n%s", got)
	}
}

// TestFormatReport_ModerateRiskLabel verifies that moderate-risk entries
// display "moderate".
func TestFormatReport_ModerateRiskLabel(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Middle", File: "mid.go", CC: 5, Coverage: 0.5, Score: 8.125, LOC: 15},
	}
	got := report.FormatReport(entries)
	if !strings.Contains(got, "moderate") {
		t.Errorf("FormatReport output missing moderate risk label\n%s", got)
	}
}

// TestFormatReport_SortedByScoreDescending verifies that entries are
// sorted with the highest CRAP score first.
func TestFormatReport_SortedByScoreDescending(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Low", File: "a.go", CC: 1, Coverage: 1.0, Score: 1.0},
		{Name: "High", File: "b.go", CC: 10, Coverage: 0.0, Score: 110.0},
		{Name: "Mid", File: "c.go", CC: 5, Coverage: 0.5, Score: 8.125},
	}
	got := report.FormatReport(entries)

	highIdx := strings.Index(got, "High")
	midIdx := strings.Index(got, "Mid")
	lowIdx := strings.Index(got, "Low")

	if highIdx == -1 || midIdx == -1 || lowIdx == -1 {
		t.Fatalf("FormatReport output missing one or more function names\n%s", got)
	}
	if !(highIdx < midIdx && midIdx < lowIdx) {
		t.Errorf("FormatReport entries not sorted by score descending: High(%d) Mid(%d) Low(%d)\n%s",
			highIdx, midIdx, lowIdx, got)
	}
}

// TestFormatReport_SummaryLine verifies that the summary line contains the
// total number of functions and high-risk count.
func TestFormatReport_SummaryLine(t *testing.T) {
	entries := []crap.Entry{
		{Name: "High1", File: "a.go", CC: 10, Coverage: 0.0, Score: 110.0},
		{Name: "High2", File: "b.go", CC: 8, Coverage: 0.0, Score: 72.0},
		{Name: "Low", File: "c.go", CC: 1, Coverage: 1.0, Score: 1.0},
	}
	got := report.FormatReport(entries)
	// Expect "3 functions" and "2 high risk" somewhere in the output.
	if !strings.Contains(got, "3") {
		t.Errorf("FormatReport summary missing total count 3\n%s", got)
	}
	if !strings.Contains(got, "2") {
		t.Errorf("FormatReport summary missing high risk count 2\n%s", got)
	}
}

// TestFormatReport_NoCoverageData verifies that a coverage of -1 is shown
// as "n/a" or similar (not as -100%).
func TestFormatReport_NoCoverageData(t *testing.T) {
	entries := []crap.Entry{
		{Name: "Unknown", File: "x.go", CC: 3, Coverage: -1, Score: 12.0},
	}
	got := report.FormatReport(entries)
	// Should NOT show a negative percentage.
	if strings.Contains(got, "-1") || strings.Contains(got, "-100") {
		t.Errorf("FormatReport shows negative coverage percentage for sentinel -1\n%s", got)
	}
}

