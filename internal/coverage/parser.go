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
package coverage

import (
	"go/token"

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
//
// TODO(prompt-3): implement coverage profile parsing.
func LoadProfile(filename string) (Profile, error) {
	// Stub implementation — full logic is added in Prompt 3.
	_ = filename
	return Profile{}, nil
}

// MapCoverage maps coverage data from profile onto the provided functions and
// returns a map from function key ("File:Name") to coverage fraction (0.0–1.0).
//
// Functions with no overlapping coverage blocks are mapped to the sentinel
// value −1, indicating "no data available".
//
// fset is the token.FileSet used when the source files were parsed; it is
// used to convert block line ranges to precise token positions.
//
// TODO(prompt-3): implement coverage mapping logic.
func MapCoverage(
	profile Profile,
	functions []*complexity.Function,
	fset *token.FileSet,
) map[string]float64 {
	// Stub implementation — full logic is added in Prompt 3.
	_ = profile
	_ = functions
	_ = fset
	return map[string]float64{}
}
