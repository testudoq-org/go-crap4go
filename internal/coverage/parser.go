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

	var p Profile
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if lineNum == 1 {
			// First line must be "mode: <mode>"
			mode, err := parseModeLine(line)
			if err != nil {
				return Profile{}, fmt.Errorf("coverage: %s:%d: %w", filename, lineNum, err)
			}
			p.Mode = mode
			continue
		}

		b, err := parseBlockLine(line)
		if err != nil {
			return Profile{}, fmt.Errorf("coverage: %s:%d: %w", filename, lineNum, err)
		}
		p.Blocks = append(p.Blocks, b)
	}

	if err := scanner.Err(); err != nil {
		return Profile{}, fmt.Errorf("coverage: scan %q: %w", filename, err)
	}
	if p.Mode == "" {
		return Profile{}, fmt.Errorf("coverage: %q: empty or missing mode line", filename)
	}

	return p, nil
}

// MapCoverage maps coverage data from profile onto the provided functions and
// returns a map from function key ("File:Name") to coverage fraction (0.0–1.0).
//
// Functions with no overlapping coverage blocks are mapped to the sentinel
// value −1, indicating "no data available".
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

	for _, fn := range functions {
		key := fn.File + ":" + fn.Name
		total, covered := 0, 0

		for i := range profile.Blocks {
			b := &profile.Blocks[i]
			if !fileMatches(b.File, fn.File) {
				continue
			}
			if b.StartLine >= fn.StartLine && b.EndLine <= fn.EndLine {
				total += b.NumStmts
				if b.Count > 0 {
					covered += b.NumStmts
				}
			}
		}

		if total == 0 {
			result[key] = -1
		} else {
			result[key] = float64(covered) / float64(total)
		}
	}

	return result
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
	// Split on the last colon before the numeric range to isolate the file path.
	// File paths themselves may contain colons on Windows; the range always
	// starts after the final colon that precedes a digit.
	colonIdx := strings.LastIndex(line, ":")
	if colonIdx < 0 {
		return Block{}, fmt.Errorf("missing colon in block line: %q", line)
	}
	filePath := line[:colonIdx]
	rest := line[colonIdx+1:]

	// rest = "<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>"
	fields := strings.Fields(rest)
	if len(fields) != 3 {
		return Block{}, fmt.Errorf("expected 3 fields after file path, got %d in %q", len(fields), line)
	}

	ranges := strings.Split(fields[0], ",")
	if len(ranges) != 2 {
		return Block{}, fmt.Errorf("expected start,end range in %q", fields[0])
	}

	startLine, startCol, err := parseLineCol(ranges[0])
	if err != nil {
		return Block{}, fmt.Errorf("parsing start position: %w", err)
	}
	endLine, endCol, err := parseLineCol(ranges[1])
	if err != nil {
		return Block{}, fmt.Errorf("parsing end position: %w", err)
	}

	numStmts, err := strconv.Atoi(fields[1])
	if err != nil {
		return Block{}, fmt.Errorf("parsing numStmts %q: %w", fields[1], err)
	}
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return Block{}, fmt.Errorf("parsing count %q: %w", fields[2], err)
	}

	return Block{
		File:      filePath,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmts:  numStmts,
		Count:     count,
	}, nil
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

