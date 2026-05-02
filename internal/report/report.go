// Package report formats crap4go analysis results for human-readable
// terminal output.
//
// Output style mirrors crap4js: a fixed-width table with columns for
// function name, file, cyclomatic complexity, coverage percentage, CRAP
// score, and risk level, followed by a summary line.
//
// Example output:
//
//	FUNCTION                      FILE                          CC   COV%   CRAP  RISK
//	──────────────────────────────────────────────────────────────────────────────────
//	Extract                       complexity/extractor.go        7   42.0   41.2  HIGH
//	ParseProfile                  coverage/parser.go             3   87.5    3.5  low
//	Score                         crap/crap.go                   1  100.0    1.0  low
//	──────────────────────────────────────────────────────────────────────────────────
//	3 functions analysed • 1 high risk • threshold 30
package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-crap4go/crap4go/internal/crap"
)

// column widths for the fixed-width table
const (
	colFunc = 30
	colFile = 30
	colCC   = 4
	colCov  = 7
	colCRAP = 7
	colRisk = 8
)

// separator is drawn between the header and the rows, and again before the
// summary line.
var separator = strings.Repeat("─", colFunc+colFile+colCC+colCov+colCRAP+colRisk+5)

// FormatReport formats a slice of crap.Entry values as a human-readable table
// and returns the result as a string suitable for writing to stdout.
//
// Entries are sorted by CRAP score descending so the riskiest functions
// appear at the top. High-risk entries (score ≥ 30) show "HIGH" in the risk
// column; lower-risk entries use lowercase labels.
//
// Coverage of −1 (no data available) is displayed as "n/a" rather than a
// negative percentage.
func FormatReport(entries []crap.Entry) string {
	// Work on a copy so we do not mutate the caller's slice.
	sorted := make([]crap.Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	var sb strings.Builder

	// Header row.
	fmt.Fprintf(&sb, "%-*s %-*s %*s %*s %*s  %s\n",
		colFunc, "FUNCTION",
		colFile, "FILE",
		colCC, "CC",
		colCov, "COV%",
		colCRAP, "CRAP",
		"RISK",
	)
	sb.WriteString(separator)
	sb.WriteByte('\n')

	// Data rows.
	highCount := 0
	for _, e := range sorted {
		risk := crap.RiskLevel(e.Score)
		riskLabel := risk
		if risk == "high" {
			riskLabel = "HIGH"
			highCount++
		}

		covStr := formatCoverage(e.Coverage)

		fmt.Fprintf(&sb, "%-*s %-*s %*d %*s %*.1f  %s\n",
			colFunc, truncate(e.Name, colFunc),
			colFile, truncate(e.File, colFile),
			colCC, e.CC,
			colCov, covStr,
			colCRAP, e.Score,
			riskLabel,
		)
	}

	// Summary line.
	sb.WriteString(separator)
	sb.WriteByte('\n')
	fmt.Fprintf(&sb, "%d functions analysed • %d high risk\n", len(sorted), highCount)

	return sb.String()
}

// formatCoverage converts a coverage fraction (0.0–1.0) to a display string.
// A sentinel value of −1 (no data) is shown as "n/a".
func formatCoverage(cov float64) string {
	if cov < 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", cov*100)
}

// truncate shortens s to at most n runes, appending "…" if truncation occurs.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

