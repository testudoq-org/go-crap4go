package config_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-crap4go/crap4go/internal/config"
)

// configTestdata returns the absolute path to a named file under testdata/config/.
func configTestdata(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	return filepath.Join(root, "testdata", "config", name)
}

// ---------------------------------------------------------------------------
// LoadFile — error cases
// ---------------------------------------------------------------------------

// TestLoadFile_NonExistent verifies that a missing file returns an error.
func TestLoadFile_NonExistent(t *testing.T) {
	_, err := config.LoadFile("/does/not/exist.json")
	if err == nil {
		t.Error("LoadFile returned nil error for non-existent file; want error")
	}
}

// TestLoadFile_UnsupportedExtension verifies that an unsupported file
// extension returns an error describing supported formats.
func TestLoadFile_UnsupportedExtension(t *testing.T) {
	_, err := config.LoadFile(configTestdata(t, "config.yaml"))
	if err == nil {
		t.Error("LoadFile returned nil error for unsupported extension; want error")
	}
}

// TestLoadFile_InvalidJSON verifies that malformed JSON content returns an error.
func TestLoadFile_InvalidJSON(t *testing.T) {
	_, err := config.LoadFile(configTestdata(t, "invalid.json"))
	if err == nil {
		t.Error("LoadFile returned nil error for invalid JSON; want error")
	}
}

// TestLoadFile_InvalidTOML verifies that malformed TOML content returns an error.
func TestLoadFile_InvalidTOML(t *testing.T) {
	_, err := config.LoadFile(configTestdata(t, "invalid.toml"))
	if err == nil {
		t.Error("LoadFile returned nil error for invalid TOML; want error")
	}
}

// ---------------------------------------------------------------------------
// LoadFile — minimal configs
// ---------------------------------------------------------------------------

// TestLoadFile_MinimalJSON verifies that a minimal JSON config file sets only
// the threshold field.
func TestLoadFile_MinimalJSON(t *testing.T) {
	fc, err := config.LoadFile(configTestdata(t, "minimal.json"))
	if err != nil {
		t.Fatalf("LoadFile minimal.json: %v", err)
	}
	if fc.Threshold == nil {
		t.Fatal("Threshold is nil; want non-nil")
	}
	if *fc.Threshold != 20 {
		t.Errorf("Threshold = %d; want 20", *fc.Threshold)
	}
	// Fields not in the file must remain nil.
	if fc.CoverProfile != nil {
		t.Errorf("CoverProfile = %v; want nil (not set in file)", fc.CoverProfile)
	}
}

// TestLoadFile_MinimalTOML verifies that a minimal TOML config file sets only
// the threshold field.
func TestLoadFile_MinimalTOML(t *testing.T) {
	fc, err := config.LoadFile(configTestdata(t, "minimal.toml"))
	if err != nil {
		t.Fatalf("LoadFile minimal.toml: %v", err)
	}
	if fc.Threshold == nil {
		t.Fatal("Threshold is nil; want non-nil")
	}
	if *fc.Threshold != 20 {
		t.Errorf("Threshold = %d; want 20", *fc.Threshold)
	}
	if fc.CoverProfile != nil {
		t.Errorf("CoverProfile = %v; want nil (not set in file)", fc.CoverProfile)
	}
}

// ---------------------------------------------------------------------------
// LoadFile — full configs (all supported fields)
// ---------------------------------------------------------------------------

// TestLoadFile_FullJSON verifies that all supported fields are loaded from JSON.
func TestLoadFile_FullJSON(t *testing.T) {
	fc, err := config.LoadFile(configTestdata(t, "full.json"))
	if err != nil {
		t.Fatalf("LoadFile full.json: %v", err)
	}
	assertFullFileConfig(t, fc)
}

// TestLoadFile_FullTOML verifies that all supported fields are loaded from TOML.
func TestLoadFile_FullTOML(t *testing.T) {
	fc, err := config.LoadFile(configTestdata(t, "full.toml"))
	if err != nil {
		t.Fatalf("LoadFile full.toml: %v", err)
	}
	assertFullFileConfig(t, fc)
}

func assertFullFileConfig(t *testing.T, fc config.FileConfig) {
	t.Helper()
	if fc.CoverProfile == nil || *fc.CoverProfile != "coverage.out" {
		t.Errorf("CoverProfile = %v; want \"coverage.out\"", fc.CoverProfile)
	}
	if fc.Threshold == nil || *fc.Threshold != 25 {
		t.Errorf("Threshold = %v; want 25", fc.Threshold)
	}
	if fc.ShowAll == nil || !*fc.ShowAll {
		t.Errorf("ShowAll = %v; want true", fc.ShowAll)
	}
	if len(fc.Paths) != 2 {
		t.Errorf("Paths len = %d; want 2; got %v", len(fc.Paths), fc.Paths)
	}
}

// ---------------------------------------------------------------------------
// ApplyFile — precedence rules
// ---------------------------------------------------------------------------

// TestApplyFile_CLITakesPrecedence verifies that an explicitly set CLI flag
// overrides the corresponding value from the config file.
func TestApplyFile_CLITakesPrecedence(t *testing.T) {
	threshold := 25
	fc := config.FileConfig{Threshold: &threshold}
	cfg := config.Config{Threshold: 30} // as if CLI set --threshold=30
	changed := map[string]bool{"threshold": true}

	result := config.ApplyFile(cfg, fc, changed)
	if result.Threshold != 30 {
		t.Errorf("Threshold = %d; want 30 (CLI value preserved over file)", result.Threshold)
	}
}

// TestApplyFile_FileValueApplied verifies that a file value is applied when
// the corresponding CLI flag was not set.
func TestApplyFile_FileValueApplied(t *testing.T) {
	threshold := 25
	fc := config.FileConfig{Threshold: &threshold}
	cfg := config.Config{Threshold: 30} // default from cobra, not user-supplied
	changed := map[string]bool{}        // no CLI flags changed

	result := config.ApplyFile(cfg, fc, changed)
	if result.Threshold != 25 {
		t.Errorf("Threshold = %d; want 25 (file value applied)", result.Threshold)
	}
}

// TestApplyFile_NilFieldNotApplied verifies that a nil pointer field in
// FileConfig does not change the Config value.
func TestApplyFile_NilFieldNotApplied(t *testing.T) {
	fc := config.FileConfig{} // all nil
	cfg := config.Config{Threshold: 30}
	changed := map[string]bool{}

	result := config.ApplyFile(cfg, fc, changed)
	if result.Threshold != 30 {
		t.Errorf("Threshold = %d; want 30 (nil file field must not override)", result.Threshold)
	}
}

// TestApplyFile_CoverProfileApplied verifies that CoverProfile is applied from
// the file when the CLI flag was not set.
func TestApplyFile_CoverProfileApplied(t *testing.T) {
	cp := "my.out"
	fc := config.FileConfig{CoverProfile: &cp}
	cfg := config.Config{}
	changed := map[string]bool{}

	result := config.ApplyFile(cfg, fc, changed)
	if result.CoverProfile != "my.out" {
		t.Errorf("CoverProfile = %q; want \"my.out\"", result.CoverProfile)
	}
}

// TestApplyFile_PathsApplied verifies that Paths from the file are applied
// when no path filters were provided as CLI positional arguments.
func TestApplyFile_PathsApplied(t *testing.T) {
	fc := config.FileConfig{Paths: []string{"internal/crap"}}
	cfg := config.Config{} // no CLI paths
	changed := map[string]bool{}

	result := config.ApplyFile(cfg, fc, changed)
	if len(result.Paths) != 1 || result.Paths[0] != "internal/crap" {
		t.Errorf("Paths = %v; want [internal/crap]", result.Paths)
	}
}

// TestApplyFile_CLIPathsNotOverridden verifies that CLI positional path
// arguments take precedence over Paths from the config file.
func TestApplyFile_CLIPathsNotOverridden(t *testing.T) {
	fc := config.FileConfig{Paths: []string{"internal/crap"}}
	cfg := config.Config{Paths: []string{"cmd"}} // set via CLI
	changed := map[string]bool{}

	result := config.ApplyFile(cfg, fc, changed)
	if len(result.Paths) != 1 || result.Paths[0] != "cmd" {
		t.Errorf("Paths = %v; want [cmd] (CLI paths preserved)", result.Paths)
	}
}

// TestApplyFile_ShowAllApplied verifies that the ShowAll flag is applied from
// the file when the --all flag was not set on the CLI.
func TestApplyFile_ShowAllApplied(t *testing.T) {
	v := true
	fc := config.FileConfig{ShowAll: &v}
	cfg := config.Config{ShowAll: false}
	changed := map[string]bool{}

	result := config.ApplyFile(cfg, fc, changed)
	if !result.ShowAll {
		t.Error("ShowAll = false; want true (file value applied)")
	}
}
