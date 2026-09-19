package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if cfg.Interval != 1*time.Hour {
		t.Errorf("expected default interval 1h, got %v", cfg.Interval)
	}

	if cfg.Once != false {
		t.Errorf("expected default once to be false, got %v", cfg.Once)
	}
}
