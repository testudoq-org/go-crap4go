package report_test

import (
	"testing"

	"github.com/go-crap4go/crap4go/internal/crap"
	"github.com/go-crap4go/crap4go/internal/report"
)

// TestFormatReport_EmptyEntries verifies that an empty entry list returns
// a non-empty string (at minimum a header or "no results" message).
//
// TODO(prompt-1): strengthen this test once FormatReport is implemented.
func TestFormatReport_EmptyEntries(t *testing.T) {
	t.Skip("TODO(prompt-1): implement FormatReport")
	got := report.FormatReport([]crap.Entry{})
	if got == "" {
		t.Error("FormatReport(empty) returned empty string; want at least a header")
	}
}

// TestFormatReport_ContainsFunctionName will verify that the formatted output
// contains the function name for each entry.
func TestFormatReport_ContainsFunctionName(t *testing.T) {
	t.Skip("TODO(prompt-1): implement FormatReport")
}

// TestFormatReport_ContainsCRAPScore will verify that the formatted output
// includes the CRAP score for each entry.
func TestFormatReport_ContainsCRAPScore(t *testing.T) {
	t.Skip("TODO(prompt-1): implement FormatReport")
}

// TestFormatReport_HighRiskHighlighted will verify that high-risk functions
// are visually distinguished in the output (e.g. uppercased risk label).
func TestFormatReport_HighRiskHighlighted(t *testing.T) {
	t.Skip("TODO(prompt-1): implement FormatReport")
}
