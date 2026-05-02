// Package pipeline wires together the crap4go analysis pipeline:
// source file collection, AST complexity extraction, coverage mapping,
// and CRAP score computation.
package pipeline

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/go-crap4go/crap4go/internal/complexity"
	"github.com/go-crap4go/crap4go/internal/config"
	"github.com/go-crap4go/crap4go/internal/coverage"
	"github.com/go-crap4go/crap4go/internal/crap"
)

// Analyse runs the full crap4go analysis pipeline and returns one Entry per
// function found in the analysed source tree.
//
// Pipeline:
//  1. Walk cfg.Dir (default ".") for non-test .go files, filtered by cfg.Paths.
//  2. Parse each file and extract per-function cyclomatic complexity (CC) + LOC.
//  3. If cfg.CoverProfile is set, load and map coverage blocks onto functions.
//  4. Compute CRAP score for every function.
func Analyse(cfg config.Config) ([]crap.Entry, error) {
	dir := cfg.Dir
	if dir == "" {
		dir = "."
	}

	// 1. Collect non-test Go source files.
	files, err := collectFiles(dir, cfg.Paths)
	if err != nil {
		return nil, fmt.Errorf("pipeline: collect files: %w", err)
	}

	// 2. Parse and extract complexity.
	fset := token.NewFileSet()
	var functions []*complexity.Function

	for _, absPath := range files {
		// Store a path relative to dir so it matches coverage profile suffixes.
		relPath, relErr := filepath.Rel(dir, absPath)
		if relErr != nil {
			relPath = absPath
		}
		relPath = filepath.ToSlash(relPath)

		astFile, parseErr := parser.ParseFile(fset, absPath, nil, 0)
		if parseErr != nil {
			return nil, fmt.Errorf("pipeline: parse %s: %w", relPath, parseErr)
		}

		fns, extractErr := complexity.Extract(fset, astFile, relPath)
		if extractErr != nil {
			return nil, fmt.Errorf("pipeline: extract %s: %w", relPath, extractErr)
		}
		functions = append(functions, fns...)
	}

	// 3. Load coverage profile (optional).
	var coverMap map[string]float64
	if cfg.CoverProfile != "" {
		profile, loadErr := coverage.LoadProfile(cfg.CoverProfile)
		if loadErr != nil {
			return nil, fmt.Errorf("pipeline: load coverage: %w", loadErr)
		}
		coverMap = coverage.MapCoverage(profile, functions, fset)
	}

	// 4. Build Entry slice.
	entries := make([]crap.Entry, 0, len(functions))
	for _, fn := range functions {
		cov := -1.0
		if coverMap != nil {
			key := fn.File + ":" + fn.Name
			if v, ok := coverMap[key]; ok {
				cov = v
			}
		}
		score := crap.Score(fn.CC, cov)
		entries = append(entries, crap.Entry{
			Name:     fn.Name,
			File:     fn.File,
			CC:       fn.CC,
			Coverage: cov,
			Score:    score,
			LOC:      fn.LOC,
		})
	}

	return entries, nil
}

// collectFiles walks dir recursively and returns the absolute paths of all
// non-test .go files that match at least one fragment in paths (OR semantics).
// When paths is empty all non-test .go files are returned.
//
// The following directories are always skipped:
// hidden directories (starting with "."), "vendor", and "testdata".
func collectFiles(dir string, paths []string) ([]string, error) {
	var result []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only .go files; skip test files.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Path filter: keep files whose slash-normalised relative path contains
		// at least one of the filter fragments.
		if len(paths) > 0 {
			rel, _ := filepath.Rel(dir, path)
			rel = filepath.ToSlash(rel)
			matched := false
			for _, p := range paths {
				if strings.Contains(rel, filepath.ToSlash(p)) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		result = append(result, path)
		return nil
	})

	return result, err
}

