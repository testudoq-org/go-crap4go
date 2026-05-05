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
			if cfg.ConfigFile != "" {
				fc, loadErr := config.LoadFile(cfg.ConfigFile)
				if loadErr != nil {
					return loadErr
				}
				cfg = config.ApplyFile(cfg, fc, changedFlags(cmd,
					"coverprofile", "run-tests", "no-run-tests", "threshold", "all"))
			}
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
	if err := validateFlags(cfg); err != nil {
		return err
	}
	tmpProfile, cfg, err := resolveProfile(cfg)
	if tmpProfile != "" {
		defer os.Remove(tmpProfile)
	}
	if err != nil {
		return err
	}
	return analyse(cfg)
}

// validateFlags checks for invalid flag combinations.
func validateFlags(cfg config.Config) error {
	if cfg.RunTests && cfg.NoRunTests {
		return fmt.Errorf("--run-tests and --no-run-tests are mutually exclusive")
	}
	if cfg.NoRunTests && cfg.CoverProfile == "" {
		return fmt.Errorf("--no-run-tests requires --coverprofile to be set")
	}
	return nil
}

// resolveProfile determines whether to run go test and returns the temporary
// coverage profile path (non-empty when a temp file was created by this call).
func resolveProfile(cfg config.Config) (tmpPath string, updated config.Config, err error) {
	needRun := cfg.RunTests || (cfg.CoverProfile == "" && !cfg.NoRunTests)
	if !needRun {
		return "", cfg, nil
	}
	tmp, err := runGoTests()
	if err != nil {
		return "", cfg, fmt.Errorf("running go test: %w", err)
	}
	cfg.CoverProfile = tmp
	return tmp, cfg, nil
}

// analyse runs the pipeline, prints the report, and checks the threshold.
func analyse(cfg config.Config) error {
	entries, err := pipeline.Analyse(cfg)
	if err != nil {
		return err
	}
	fmt.Print(report.FormatReport(filterEntries(entries, cfg.ShowAll)))
	if high := countHighRisk(entries, cfg.Threshold); high > 0 {
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

// changedFlags builds a set of flag names that were explicitly set on the
// command line, allowing config file values to be applied only for flags
// that the user did not provide.
func changedFlags(cmd *cobra.Command, names ...string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, name := range names {
		m[name] = cmd.Flags().Changed(name)
	}
	return m
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
