// Package pipeline wires together the crap4go analysis pipeline:
// source file collection, AST complexity extraction, coverage mapping,
// and CRAP score computation.
package pipeline

import (
	"fmt"
	"go/build"
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

	files, err := collectFiles(dir, cfg.Paths)
	if err != nil {
		return nil, fmt.Errorf("pipeline: collect files: %w", err)
	}

	fset := token.NewFileSet()
	functions, err := parseFiles(dir, files, fset)
	if err != nil {
		return nil, err
	}

	coverMap, err := loadCoverageMap(cfg.CoverProfile, functions, fset)
	if err != nil {
		return nil, err
	}

	return buildEntries(functions, coverMap), nil
}

// parseFiles parses each .go file in files and extracts function descriptors.
func parseFiles(dir string, files []string, fset *token.FileSet) ([]*complexity.Function, error) {
	var functions []*complexity.Function
	for _, absPath := range files {
		fns, err := parseOneFile(dir, absPath, fset)
		if err != nil {
			return nil, err
		}
		functions = append(functions, fns...)
	}
	return functions, nil
}

// parseOneFile parses a single Go source file and returns its function list.
func parseOneFile(dir, absPath string, fset *token.FileSet) ([]*complexity.Function, error) {
	relPath, relErr := filepath.Rel(dir, absPath)
	if relErr != nil {
		relPath = absPath
	}
	relPath = filepath.ToSlash(relPath)

	astFile, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("pipeline: parse %s: %w", relPath, err)
	}

	fns, err := complexity.Extract(fset, astFile, relPath)
	if err != nil {
		return nil, fmt.Errorf("pipeline: extract %s: %w", relPath, err)
	}
	return fns, nil
}

// loadCoverageMap loads the coverage profile (if any) and maps it onto fns.
// Returns nil, nil when profile is empty (no coverage data available).
func loadCoverageMap(profile string, functions []*complexity.Function, fset *token.FileSet) (map[string]float64, error) {
	if profile == "" {
		return nil, nil
	}
	p, err := coverage.LoadProfile(profile)
	if err != nil {
		return nil, fmt.Errorf("pipeline: load coverage: %w", err)
	}
	return coverage.MapCoverage(p, functions, fset), nil
}

// buildEntries converts function descriptors and coverage data into crap.Entry values.
func buildEntries(functions []*complexity.Function, coverMap map[string]float64) []crap.Entry {
	entries := make([]crap.Entry, 0, len(functions))
	for _, fn := range functions {
		cov := coverageFor(fn, coverMap)
		entries = append(entries, crap.Entry{
			Name:     fn.Name,
			File:     fn.File,
			CC:       fn.CC,
			Coverage: cov,
			Score:    crap.Score(fn.CC, cov),
			LOC:      fn.LOC,
		})
	}
	return entries
}

// coverageFor looks up the coverage fraction for fn in coverMap.
// Returns -1.0 when no coverage data is available.
func coverageFor(fn *complexity.Function, coverMap map[string]float64) float64 {
	if coverMap == nil {
		return -1.0
	}
	if v, ok := coverMap[fn.File+":"+fn.Name]; ok {
		return v
	}
	return -1.0
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
			return skipDirDecision(path, dir, d)
		}
		if !isGoSourceFile(path) || !matchesFilter(path, dir, paths) {
			return nil
		}
		// Skip files whose build constraints are not satisfied by the current
		// build context (e.g. //go:build ignore, OS/arch constraints).
		match, err := build.Default.MatchFile(filepath.Dir(path), filepath.Base(path))
		if err != nil || !match {
			return nil
		}
		result = append(result, path)
		return nil
	})

	return result, err
}

// skipDirDecision returns filepath.SkipDir for directories that should be
// excluded from the walk (hidden dirs, vendor, testdata). The root dir itself
// is never skipped.
func skipDirDecision(path, rootDir string, d fs.DirEntry) error {
	if path == rootDir {
		return nil
	}
	name := d.Name()
	if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" {
		return filepath.SkipDir
	}
	return nil
}

// isGoSourceFile reports whether path is a non-test Go source file.
func isGoSourceFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

// matchesFilter reports whether path (relative to dir) contains at least one
// of the filter fragments. Returns true when paths is empty (no filter).
func matchesFilter(path, dir string, paths []string) bool {
	if len(paths) == 0 {
		return true
	}
	rel, _ := filepath.Rel(dir, path)
	rel = filepath.ToSlash(rel)
	for _, p := range paths {
		if strings.Contains(rel, filepath.ToSlash(p)) {
			return true
		}
	}
	return false
}
