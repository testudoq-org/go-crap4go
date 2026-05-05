// Package coverage parses Go coverage profiles (coverage.out) and maps
// per-block coverage data onto extracted function descriptors.
//
// # Coverage Profile Format
//
// Go coverage profiles use the following text format:
//
//	mode: <set|count|atomic>
//	<file>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
//
// Where count > 0 means the block was executed at least once ("set" mode)
// or N times ("count"/"atomic" mode).
//
// # Mapping Strategy
//
// For each function F with span [startLine, endLine], the coverage fraction
// is computed as:
//
//	coveredStmts / totalStmts
//
// where totalStmts is the sum of NumStmts for all blocks whose line range
// falls entirely within F's span, and coveredStmts sums only blocks with
// Count > 0.
//
// Functions with no overlapping coverage blocks receive a sentinel value of
// −1, which the crap package treats as 0% coverage.
//
// Coverage profile file paths use the full module path (e.g.
// "github.com/go-crap4go/crap4go/internal/crap/crap.go"). Function.File
// fields are workspace-relative (e.g. "internal/crap/crap.go"). MapCoverage
// matches a block to a function when the block's file path has a suffix that
// equals the function's file path (after normalising separators).
package coverage

import (
	"bufio"
	"fmt"
	"go/token"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/go-crap4go/crap4go/internal/complexity"
)

// Block represents a single coverage block parsed from a Go coverage profile.
type Block struct {
	// File is the module-relative source path recorded in the profile
	// (e.g. "github.com/go-crap4go/crap4go/internal/crap/crap.go").
	File string

	// StartLine and EndLine are the 1-based line numbers bounding the block.
	StartLine int
	EndLine   int

	// StartCol and EndCol are the 1-based column offsets bounding the block.
	StartCol int
	EndCol   int

	// NumStmts is the number of statements covered by this block.
	NumStmts int

	// Count is the execution count recorded by the test run.
	// In "set" mode this is 0 or 1; in "count"/"atomic" mode it is the
	// actual invocation count.
	Count int
}

// Profile holds the parsed contents of a Go coverage profile file.
type Profile struct {
	// Mode is one of "set", "count", or "atomic" as declared in the first
	// line of the coverage profile.
	Mode string

	// Blocks contains all coverage blocks in the order they appear in the
	// profile file.
	Blocks []Block
}

// LoadProfile reads and parses a Go coverage profile from filename.
//
// It returns a non-nil error when the file cannot be opened, is empty, or
// contains lines that do not conform to the expected format.
func LoadProfile(filename string) (Profile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Profile{}, fmt.Errorf("coverage: open %q: %w", filename, err)
	}
	defer f.Close()
	return scanProfile(filename, f)
}

// scanProfile reads coverage lines from r and builds a Profile.
func scanProfile(filename string, r io.Reader) (Profile, error) {
	var p Profile
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if err := processLine(filename, lineNum, strings.TrimSpace(scanner.Text()), &p); err != nil {
			return Profile{}, err
		}
	}

	if err := scanner.Err(); err != nil {
		return Profile{}, fmt.Errorf("coverage: scan %q: %w", filename, err)
	}
	if p.Mode == "" {
		return Profile{}, fmt.Errorf("coverage: %q: empty or missing mode line", filename)
	}
	return p, nil
}

// processLine handles a single trimmed line from the coverage profile.
func processLine(filename string, lineNum int, line string, p *Profile) error {
	if line == "" {
		return nil
	}
	if lineNum == 1 {
		mode, err := parseModeLine(line)
		if err != nil {
			return fmt.Errorf("coverage: %s:%d: %w", filename, lineNum, err)
		}
		p.Mode = mode
		return nil
	}
	b, err := parseBlockLine(line)
	if err != nil {
		return fmt.Errorf("coverage: %s:%d: %w", filename, lineNum, err)
	}
	p.Blocks = append(p.Blocks, b)
	return nil
}

// MapCoverage maps coverage data from profile onto the provided functions and
// returns a map from function key ("File:Name") to coverage fraction (0.0–1.0).
//
// Functions with no overlapping coverage blocks are mapped to the sentinel
// value −1, indicating "no data available".
//
// When a closure's line range is nested inside an outer function, coverage
// blocks that fall within the closure are attributed only to the innermost
// matching function. This prevents double-counting and gives accurate
// coverage fractions for both the closure and its enclosing function.
//
// fset is accepted for API compatibility but is not required for the current
// line-based matching strategy.
func MapCoverage(
	profile Profile,
	functions []*complexity.Function,
	fset *token.FileSet,
) map[string]float64 {
	result := make(map[string]float64, len(functions))
	_ = fset // reserved for future position-based matching

	if len(functions) == 0 {
		return result
	}

	// Sort a working copy by span length ascending so the innermost (shortest
	// span) function is considered first for each block.
	sorted := make([]*complexity.Function, len(functions))
	copy(sorted, functions)
	sortBySpan(sorted)

	// Assign each block to the innermost function whose file and line range
	// contain it. Track already-claimed blocks by index so outer functions
	// cannot double-count blocks belonging to nested closures.
	type tally struct{ total, covered int }
	tallies := make(map[string]*tally, len(functions))
	for _, fn := range sorted {
		tallies[fn.File+":"+fn.Name] = &tally{}
	}
	claimed := make([]bool, len(profile.Blocks))

	for _, fn := range sorted {
		key := fn.File + ":" + fn.Name
		t := tallies[key]
		for i := range profile.Blocks {
			if claimed[i] {
				continue
			}
			b := &profile.Blocks[i]
			if !fileMatches(b.File, fn.File) || !blockInRange(b, fn) {
				continue
			}
			claimed[i] = true
			t.total += b.NumStmts
			if b.Count > 0 {
				t.covered += b.NumStmts
			}
		}
	}

	for _, fn := range functions {
		key := fn.File + ":" + fn.Name
		t := tallies[key]
		if t.total == 0 {
			result[key] = -1
		} else {
			result[key] = float64(t.covered) / float64(t.total)
		}
	}

	return result
}

// sortBySpan sorts functions by span length (EndLine-StartLine) ascending,
// so innermost (shortest) functions are processed first during block claiming.
func sortBySpan(fns []*complexity.Function) {
	// insertion sort — function lists are typically small
	for i := 1; i < len(fns); i++ {
		for j := i; j > 0; j-- {
			ai := fns[j].EndLine - fns[j].StartLine
			aj := fns[j-1].EndLine - fns[j-1].StartLine
			if ai < aj {
				fns[j], fns[j-1] = fns[j-1], fns[j]
			} else {
				break
			}
		}
	}
}

// blockInRange reports whether b falls entirely within fn's line span.
func blockInRange(b *Block, fn *complexity.Function) bool {
	return b.StartLine >= fn.StartLine && b.EndLine <= fn.EndLine
}

// ---------------------------------------------------------------------------
// parsing helpers
// ---------------------------------------------------------------------------

// parseModeLine parses the first line of a coverage profile ("mode: set").
func parseModeLine(line string) (string, error) {
	const prefix = "mode: "
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("expected \"mode: <mode>\", got %q", line)
	}
	mode := strings.TrimPrefix(line, prefix)
	switch mode {
	case "set", "count", "atomic":
		return mode, nil
	default:
		return "", fmt.Errorf("unrecognised coverage mode %q", mode)
	}
}

// parseBlockLine parses a single block line of the form:
//
//	<file>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>
func parseBlockLine(line string) (Block, error) {
	filePath, rest, err := splitFileAndRest(line)
	if err != nil {
		return Block{}, err
	}
	return parseBlockRest(filePath, rest, line)
}

// splitFileAndRest splits a block line at the last colon to separate the file
// path from the numeric range and count fields.
func splitFileAndRest(line string) (filePath, rest string, err error) {
	colonIdx := strings.LastIndex(line, ":")
	if colonIdx < 0 {
		return "", "", fmt.Errorf("missing colon in block line: %q", line)
	}
	return line[:colonIdx], line[colonIdx+1:], nil
}

// parseBlockRest parses the "startLine.col,endLine.col numStmts count" portion
// of a block line, given the already-isolated file path.
func parseBlockRest(filePath, rest, line string) (Block, error) {
	fields := strings.Fields(rest)
	if len(fields) != 3 {
		return Block{}, fmt.Errorf("expected 3 fields after file path, got %d in %q", len(fields), line)
	}

	start, end, err := parsePositions(fields[0])
	if err != nil {
		return Block{}, err
	}

	numStmts, count, err := parseCounts(fields[1], fields[2])
	if err != nil {
		return Block{}, err
	}

	return Block{
		File:      filePath,
		StartLine: start[0], StartCol: start[1],
		EndLine: end[0], EndCol: end[1],
		NumStmts: numStmts, Count: count,
	}, nil
}

// parsePositions parses "startLine.startCol,endLine.endCol" into two [2]int.
func parsePositions(rangeStr string) (start, end [2]int, err error) {
	ranges := strings.Split(rangeStr, ",")
	if len(ranges) != 2 {
		return [2]int{}, [2]int{}, fmt.Errorf("expected start,end range in %q", rangeStr)
	}
	sl, sc, err := parseLineCol(ranges[0])
	if err != nil {
		return [2]int{}, [2]int{}, fmt.Errorf("parsing start position: %w", err)
	}
	el, ec, err := parseLineCol(ranges[1])
	if err != nil {
		return [2]int{}, [2]int{}, fmt.Errorf("parsing end position: %w", err)
	}
	return [2]int{sl, sc}, [2]int{el, ec}, nil
}

// parseCounts parses the numStmts and count string fields into integers.
func parseCounts(numStmtsStr, countStr string) (numStmts, count int, err error) {
	numStmts, err = strconv.Atoi(numStmtsStr)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing numStmts %q: %w", numStmtsStr, err)
	}
	count, err = strconv.Atoi(countStr)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing count %q: %w", countStr, err)
	}
	return numStmts, count, nil
}

// parseLineCol parses "line.col" into two ints.
func parseLineCol(s string) (line, col int, err error) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected line.col, got %q", s)
	}
	line, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing line number %q: %w", parts[0], err)
	}
	col, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing column %q: %w", parts[1], err)
	}
	return line, col, nil
}

// ---------------------------------------------------------------------------
// path matching
// ---------------------------------------------------------------------------

// fileMatches reports whether blockFile (full module path) corresponds to
// funcFile (workspace-relative path). It uses a suffix match after
// normalising path separators.
func fileMatches(blockFile, funcFile string) bool {
	// Normalise to forward slashes for cross-platform safety.
	bf := strings.ReplaceAll(blockFile, "\\", "/")
	ff := strings.ReplaceAll(funcFile, "\\", "/")

	return bf == ff || strings.HasSuffix(bf, "/"+ff)
}
