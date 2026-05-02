// Package report formats crap4go analysis results for human-readable
// terminal output.
//
// Output style mirrors crap4js: a fixed-width table with columns for
// function name, file, cyclomatic complexity, coverage percentage, CRAP
// score, and risk level, followed by a summary line.
//
// Example output:
//
//	FUNCTION                      FILE                      CC   COV%   CRAP  RISK
//	────────────────────────────────────────────────────────────────────────────
//	Extract                       complexity/extractor.go    7   42.0   41.2  HIGH
//	ParseProfile                  coverage/parser.go         3   87.5    3.5  low
//	Score                         crap/crap.go               1  100.0    1.0  low
//	────────────────────────────────────────────────────────────────────────────
//	3 functions analysed • 1 high risk • threshold 30
package report

import (
	"github.com/go-crap4go/crap4go/internal/crap"
)

// FormatReport formats a slice of crap.Entry values as a human-readable table
// and returns the result as a string suitable for writing to stdout.
//
// Entries are sorted by CRAP score descending so the riskiest functions
// appear at the top.
//
// TODO(prompt-1): implement full table formatting.
func FormatReport(entries []crap.Entry) string {
	// Stub implementation — full logic is added in Prompt 1.
	_ = entries
	return ""
}
