// Package crap implements the CRAP (Change Risk Anti-Patterns) formula,
// risk classification, and the core data types shared across crap4go.
//
// # Formula
//
//	CRAP = CC² × (1 − coverage)³ + CC
//
// Where:
//   - CC       is the cyclomatic complexity of the function (integer ≥ 1)
//   - coverage is the fraction of statements covered by tests (0.0–1.0)
//
// # Risk thresholds (mirror crap4js)
//
//	score < 5   → low
//	5 ≤ score < 30  → moderate
//	score ≥ 30  → high
//
// A coverage value of −1 is a sentinel meaning "no data available"; it is
// treated identically to 0% coverage (entirely uncovered).
package crap

import "math"

// Entry holds the analysis result for a single Go function, method, or closure.
type Entry struct {
	// Name is the fully qualified display name.
	//   Named function:   "FuncName"
	//   Method:           "ReceiverType.MethodName"
	//   Closure:          "<anonymous:line>"
	Name string

	// File is the workspace-relative path to the source file.
	File string

	// CC is the cyclomatic complexity of the function (base value is 1).
	CC int

	// Coverage is the fraction of statements covered by tests (0.0–1.0).
	// A value of -1 signals that no coverage data is available for this
	// function; it is treated as 0% coverage when computing the score.
	Coverage float64

	// Score is the computed CRAP score.
	Score float64

	// LOC is the number of non-blank, non-comment source lines in the
	// function body, derived from the token.FileSet positions.
	LOC int
}

// Score computes the CRAP score for a function.
//
// cc must be ≥ 1. coverage must be in the range 0.0–1.0, or −1 to indicate
// that no coverage data is available (treated identically to 0.0).
//
// The formula is: CC² × (1 − coverage)³ + CC
func Score(cc int, coverage float64) float64 {
	if coverage < 0 {
		coverage = 0
	}
	c := float64(cc)
	return math.Pow(c, 2)*math.Pow(1-coverage, 3) + c
}

// RiskLevel classifies a CRAP score into a named risk category.
//
// Thresholds match those used by crap4js:
//
//	score < 5        → "low"
//	5 ≤ score < 30   → "moderate"
//	score ≥ 30       → "high"
func RiskLevel(score float64) string {
	switch {
	case score < 5:
		return "low"
	case score < 30:
		return "moderate"
	default:
		return "high"
	}
}
