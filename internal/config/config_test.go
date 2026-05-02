package config_test

import (
	"testing"

	"github.com/go-crap4go/crap4go/internal/config"
)

func TestDefault_Threshold(t *testing.T) {
	cfg := config.Default()
	if cfg.Threshold != 30 {
		t.Errorf("Default().Threshold = %d; want 30", cfg.Threshold)
	}
}

func TestDefault_ZeroValues(t *testing.T) {
	cfg := config.Default()

	if cfg.CoverProfile != "" {
		t.Errorf("Default().CoverProfile = %q; want empty string", cfg.CoverProfile)
	}
	if cfg.RunTests {
		t.Error("Default().RunTests = true; want false")
	}
	if cfg.NoRunTests {
		t.Error("Default().NoRunTests = true; want false")
	}
	if cfg.ShowAll {
		t.Error("Default().ShowAll = true; want false")
	}
	if len(cfg.Paths) != 0 {
		t.Errorf("Default().Paths = %v; want empty slice", cfg.Paths)
	}
}
