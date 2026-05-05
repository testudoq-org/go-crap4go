package crap_test

import (
	"math"
	"testing"

	"github.com/go-crap4go/crap4go/internal/crap"
)

// ---------------------------------------------------------------------------
// Score
// ---------------------------------------------------------------------------

func TestScore_PerfectCoverage(t *testing.T) {
	// With 100% coverage the uncovered term vanishes: CC² × 0³ + CC = CC.
	got := crap.Score(3, 1.0)
	want := 3.0
	if !almostEqual(got, want) {
		t.Errorf("Score(3, 1.0) = %f; want %f", got, want)
	}
}

func TestScore_ZeroCoverage(t *testing.T) {
	// With 0% coverage: CC² × 1³ + CC = CC² + CC.
	got := crap.Score(3, 0.0)
	want := 12.0 // 9 + 3
	if !almostEqual(got, want) {
		t.Errorf("Score(3, 0.0) = %f; want %f", got, want)
	}
}

func TestScore_NoCoverageData(t *testing.T) {
	// Sentinel value -1 is treated identically to 0% coverage.
	withSentinel := crap.Score(3, -1)
	withZero := crap.Score(3, 0.0)
	if !almostEqual(withSentinel, withZero) {
		t.Errorf("Score(3,-1) = %f; want same as Score(3,0.0) = %f", withSentinel, withZero)
	}
}

func TestScore_PartialCoverage(t *testing.T) {
	// CC=5, coverage=0.5 → 25 × 0.125 + 5 = 3.125 + 5 = 8.125
	got := crap.Score(5, 0.5)
	want := 8.125
	if !almostEqual(got, want) {
		t.Errorf("Score(5, 0.5) = %f; want %f", got, want)
	}
}

func TestScore_CC1_FullCoverage(t *testing.T) {
	// Simplest function fully covered: CRAP = 1.
	got := crap.Score(1, 1.0)
	if !almostEqual(got, 1.0) {
		t.Errorf("Score(1, 1.0) = %f; want 1.0", got)
	}
}

func TestScore_HighComplexityNoCoverage(t *testing.T) {
	// CC=10, coverage=0 → 100 + 10 = 110 (well into "high" territory).
	got := crap.Score(10, 0.0)
	want := 110.0
	if !almostEqual(got, want) {
		t.Errorf("Score(10, 0.0) = %f; want %f", got, want)
	}
}

// ---------------------------------------------------------------------------
// RiskLevel
// ---------------------------------------------------------------------------

func TestRiskLevel_Low(t *testing.T) {
	cases := []float64{0.0, 1.0, 4.0, 4.99}
	for _, score := range cases {
		if got := crap.RiskLevel(score); got != "low" {
			t.Errorf("RiskLevel(%f) = %q; want \"low\"", score, got)
		}
	}
}

func TestRiskLevel_Moderate(t *testing.T) {
	cases := []float64{5.0, 10.0, 20.0, 29.99}
	for _, score := range cases {
		if got := crap.RiskLevel(score); got != "moderate" {
			t.Errorf("RiskLevel(%f) = %q; want \"moderate\"", score, got)
		}
	}
}

func TestRiskLevel_High(t *testing.T) {
	cases := []float64{30.0, 50.0, 110.0, 1000.0}
	for _, score := range cases {
		if got := crap.RiskLevel(score); got != "high" {
			t.Errorf("RiskLevel(%f) = %q; want \"high\"", score, got)
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

const epsilon = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}
