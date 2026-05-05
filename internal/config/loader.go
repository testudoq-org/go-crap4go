package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// FileConfig holds configuration values that may be loaded from a crap.toml
// or .crap.json file. Pointer fields allow absent entries to be distinguished
// from zero values, enabling correct CLI-over-file precedence merging.
//
// Supported keys (TOML/JSON):
//
//	coverprofile   string   — path to a Go coverage profile
//	run-tests      bool     — force go test execution
//	no-run-tests   bool     — skip go test (requires coverprofile)
//	threshold      int      — minimum CRAP score for high risk (default 30)
//	all            bool     — include low-risk functions in output
//	paths          []string — path fragment filters (OR semantics)
type FileConfig struct {
	CoverProfile *string  `json:"coverprofile" toml:"coverprofile"`
	RunTests     *bool    `json:"run-tests"    toml:"run-tests"`
	NoRunTests   *bool    `json:"no-run-tests" toml:"no-run-tests"`
	Threshold    *int     `json:"threshold"    toml:"threshold"`
	ShowAll      *bool    `json:"all"          toml:"all"`
	Paths        []string `json:"paths"        toml:"paths"`
}

// LoadFile reads a crap.toml or .crap.json configuration file and returns a
// FileConfig. Fields absent from the file remain nil (or empty for slices).
//
// The format is inferred from the file extension:
//   - .toml → TOML (key = value)
//   - .json → JSON
//
// Any other extension returns an error.
func LoadFile(path string) (FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, fmt.Errorf("config: read %q: %w", path, err)
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return parseJSON(data)
	case ".toml":
		return parseTOML(data)
	default:
		return FileConfig{}, fmt.Errorf("config: unsupported file format %q (supported: .toml, .json)", ext)
	}
}

// ApplyFile merges fc into cfg for any field whose corresponding CLI flag was
// not explicitly provided. changed maps flag names to true when the flag was
// set on the command line.
//
// Expected flag names match those registered in cmd/crap/main.go:
// "coverprofile", "run-tests", "no-run-tests", "threshold", "all".
//
// Paths (positional args) are taken from fc only when cfg.Paths is empty.
func ApplyFile(cfg Config, fc FileConfig, changed map[string]bool) Config {
	if fc.CoverProfile != nil && !changed["coverprofile"] {
		cfg.CoverProfile = *fc.CoverProfile
	}
	if fc.RunTests != nil && !changed["run-tests"] {
		cfg.RunTests = *fc.RunTests
	}
	if fc.NoRunTests != nil && !changed["no-run-tests"] {
		cfg.NoRunTests = *fc.NoRunTests
	}
	if fc.Threshold != nil && !changed["threshold"] {
		cfg.Threshold = *fc.Threshold
	}
	if fc.ShowAll != nil && !changed["all"] {
		cfg.ShowAll = *fc.ShowAll
	}
	if len(fc.Paths) > 0 && len(cfg.Paths) == 0 {
		cfg.Paths = fc.Paths
	}
	return cfg
}

// ---------------------------------------------------------------------------
// JSON parsing
// ---------------------------------------------------------------------------

func parseJSON(data []byte) (FileConfig, error) {
	var fc FileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return FileConfig{}, fmt.Errorf("config: parse JSON: %w", err)
	}
	return fc, nil
}

// ---------------------------------------------------------------------------
// TOML parsing
// ---------------------------------------------------------------------------

func parseTOML(data []byte) (FileConfig, error) {
	var fc FileConfig
	if err := toml.Unmarshal(data, &fc); err != nil {
		return FileConfig{}, fmt.Errorf("config: parse TOML: %w", err)
	}
	return fc, nil
}
