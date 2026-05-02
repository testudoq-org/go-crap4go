// Package main is the entry point for the crap4go CLI tool.
//
// It wires together configuration loading, optional test execution, AST
// complexity extraction, coverage mapping, CRAP score computation, and
// report rendering into a single cohesive pipeline.
//
// Usage:
//
//	crap [path-filters...] [flags]
//
// See the --help output for full flag documentation.
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/go-crap4go/crap4go/internal/config"
	"github.com/go-crap4go/crap4go/internal/crap"
	"github.com/go-crap4go/crap4go/internal/pipeline"
	"github.com/go-crap4go/crap4go/internal/report"
	"github.com/spf13/cobra"
)

func main() {
	if err := buildRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

// buildRootCommand constructs and returns the root cobra.Command for crap4go.
// All flag bindings are established here; actual analysis lives in run().
func buildRootCommand() *cobra.Command {
	var cfg config.Config

	root := &cobra.Command{
		Use:   "crap [path-filter...]",
		Short: "crap4go — CRAP metric analyser for Go codebases",
		Long: `crap4go computes the CRAP (Change Risk Anti-Patterns) score for every
function in a Go module and reports which ones need the most attention.

  CRAP = CC² × (1 − coverage)³ + CC

  CC       — cyclomatic complexity (branches + 1)
  coverage — fraction of statements exercised by tests (0.0–1.0)

Risk levels   score < 5 → low | 5–29 → moderate | ≥ 30 → high

Optional path-filter arguments restrict analysis to files whose paths
contain at least one of the supplied fragments (OR semantics).`,

		Example: `  # Analyse entire module (runs go test automatically)
  crap

  # Only analyse files under internal/complexity/
  crap internal/complexity

  # Use a pre-existing coverage profile, show all functions
  crap --coverprofile=coverage.out --all

  # Exit 1 if any function scores ≥ 20
  crap --threshold=20`,

		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Paths = args
			return run(cfg)
		},

		SilenceUsage: true,
	}

	f := root.Flags()
	f.StringVar(&cfg.CoverProfile, "coverprofile", "",
		"Path to an existing Go coverage profile.\n"+
			"When omitted, crap4go runs 'go test ./...' automatically\n"+
			"(unless --no-run-tests is set).")

	f.BoolVar(&cfg.RunTests, "run-tests", false,
		"Force running 'go test ./...' even when --coverprofile is provided.")

	f.BoolVar(&cfg.NoRunTests, "no-run-tests", false,
		"Do not run 'go test'. Requires --coverprofile.")

	f.IntVar(&cfg.Threshold, "threshold", 30,
		"Minimum CRAP score classified as high risk.\n"+
			"The process exits with code 1 when any function meets or\n"+
			"exceeds this value.")

	f.StringVar(&cfg.ConfigFile, "config", "",
		"Path to a crap.toml or .crap.json configuration file.")

	f.BoolVar(&cfg.ShowAll, "all", false,
		"Include low-risk functions in the report output.")

	return root
}

// run executes the full crap4go analysis pipeline for the given Config.
func run(cfg config.Config) error {
	// Validate mutually exclusive flags.
	if cfg.RunTests && cfg.NoRunTests {
		return fmt.Errorf("--run-tests and --no-run-tests are mutually exclusive")
	}
	if cfg.NoRunTests && cfg.CoverProfile == "" {
		return fmt.Errorf("--no-run-tests requires --coverprofile to be set")
	}

	// Run go test if needed to generate a coverage profile.
	needRunTests := cfg.CoverProfile == "" && !cfg.NoRunTests
	if cfg.RunTests {
		needRunTests = true
	}
	if needRunTests {
		tmpProfile, err := runGoTests()
		if err != nil {
			return fmt.Errorf("running go test: %w", err)
		}
		defer os.Remove(tmpProfile)
		cfg.CoverProfile = tmpProfile
	}

	// Analysis pipeline.
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		return err
	}

	// Filter: by default hide low-risk functions.
	displayed := filterEntries(entries, cfg.ShowAll)

	// Render report.
	fmt.Print(report.FormatReport(displayed))

	// Threshold check (against ALL entries, not just displayed).
	high := countHighRisk(entries, cfg.Threshold)
	if high > 0 {
		return fmt.Errorf("%d function(s) scored >= %d (threshold)", high, cfg.Threshold)
	}
	return nil
}

// runGoTests runs "go test -coverprofile=<tmp> ./..." in the current directory
// and returns the path to the temporary coverage file. The caller is
// responsible for deleting the file when done.
func runGoTests() (string, error) {
	tmp, err := os.CreateTemp("", "crap4go-coverage-*.out")
	if err != nil {
		return "", fmt.Errorf("create temp coverage file: %w", err)
	}
	tmp.Close()

	cmd := exec.Command("go", "test", "-coverprofile="+tmp.Name(), "./...")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("go test: %w", err)
	}
	return tmp.Name(), nil
}

// filterEntries returns entries to display. When showAll is false, low-risk
// functions (CRAP score < 5) are hidden.
func filterEntries(entries []crap.Entry, showAll bool) []crap.Entry {
	if showAll {
		return entries
	}
	result := make([]crap.Entry, 0, len(entries))
	for _, e := range entries {
		if crap.RiskLevel(e.Score) != "low" {
			result = append(result, e)
		}
	}
	return result
}

// countHighRisk returns the number of entries whose score is >= threshold.
func countHighRisk(entries []crap.Entry, threshold int) int {
	count := 0
	for _, e := range entries {
		if e.Score >= float64(threshold) {
			count++
		}
	}
	return count
}
