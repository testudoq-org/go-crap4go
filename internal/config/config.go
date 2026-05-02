// Package config defines the runtime configuration for crap4go and
// provides helpers for loading configuration from files and CLI flags.
//
// Configuration is resolved in the following priority order (highest first):
//  1. CLI flags
//  2. Config file (crap.toml / .crap.json)
//  3. Built-in defaults
package config

// Config holds all runtime configuration for the crap4go analysis pipeline.
type Config struct {
	// Paths is an optional list of path fragment filters. Only functions in
	// files whose paths contain at least one of these fragments are analysed.
	// An empty slice means analyse the entire module.
	Paths []string

	// CoverProfile is the path to an existing Go coverage profile
	// (typically named coverage.out). When empty and NoRunTests is false,
	// the tool runs "go test ./... -coverprofile=..." automatically.
	CoverProfile string

	// RunTests forces the tool to run "go test" to produce a fresh coverage
	// profile even when CoverProfile is already set.
	RunTests bool

	// NoRunTests prevents the tool from ever running "go test". When true,
	// CoverProfile must be provided; the tool exits with an error otherwise.
	NoRunTests bool

	// Threshold is the minimum CRAP score that is classified as high risk.
	// Functions at or above this value cause a non-zero exit code.
	// Default: 30 (matches crap4js).
	Threshold int

	// ConfigFile is an optional path to a crap.toml or .crap.json file.
	// CLI flags always override values loaded from the config file.
	ConfigFile string

	// ShowAll includes low-risk functions in the report when true.
	// By default only moderate and high-risk functions are shown.
	ShowAll bool
}

// Default returns a Config populated with sensible production defaults.
func Default() Config {
	return Config{
		Threshold: 30,
	}
}
