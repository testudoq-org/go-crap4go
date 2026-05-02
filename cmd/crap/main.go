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

	"github.com/go-crap4go/crap4go/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	if err := buildRootCommand().Execute(); err != nil {
		// cobra already prints the error; just set a non-zero exit code.
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

		// Silence cobra's default error formatting; we handle it ourselves.
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
//
// Pipeline steps:
//  1. Load and merge config file (if --config is set).
//  2. Optionally run 'go test ./... -coverprofile=...' to produce coverage data.
//  3. Parse all non-test .go files (respecting path filters).
//  4. Extract per-function CC and LOC via AST analysis.
//  5. Parse the coverage profile and map blocks to functions.
//  6. Compute CRAP scores.
//  7. Render and print the report.
//  8. Return an error (→ exit 1) if any function meets or exceeds the threshold.
//
// TODO(prompt-4): implement full orchestration logic.
func run(cfg config.Config) error {
	// Stub — full orchestration implemented in Prompt 4.
	fmt.Fprintf(os.Stderr,
		"crap4go: analysis pipeline not yet implemented (scaffold only)\n"+
			"  config: %+v\n", cfg)
	return nil
}
