package config

import (
	"strings"
	"testing"
)

type sharedBlock struct {
	Environment string `key:"TEST_EMBEDDED_ENVIRONMENT" default:"development"`
	Token       string `key:"TEST_EMBEDDED_TOKEN" default:""`
	Required    string `key:"TEST_EMBEDDED_REQUIRED"`
}

type plainBlock struct {
	Ignored string `key:"TEST_EMBEDDED_IGNORED" default:"untouched"`
}

type embeddingConfig struct {
	sharedBlock

	Own string `key:"TEST_EMBEDDED_OWN"`

	// Named, not embedded: must stay untouched.
	Named plainBlock
}

func TestLoadFillsEmbeddedBlock(t *testing.T) {
	t.Setenv("TEST_EMBEDDED_REQUIRED", "present")
	t.Setenv("TEST_EMBEDDED_OWN", "mine")
	t.Setenv("TEST_EMBEDDED_TOKEN", "secret-token")

	var cfg embeddingConfig
	if err := Load(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Environment != "development" {
		t.Errorf("embedded default not applied: Environment = %q", cfg.Environment)
	}

	if cfg.Token != "secret-token" {
		t.Errorf("embedded value not read: Token = %q", cfg.Token)
	}

	if cfg.Required != "present" || cfg.Own != "mine" {
		t.Errorf("loaded %+v", cfg)
	}

	if cfg.Named.Ignored != "" {
		t.Errorf("a named struct field was descended into: Named.Ignored = %q", cfg.Named.Ignored)
	}
}

func TestLoadReportsMissingEmbeddedKey(t *testing.T) {
	t.Setenv("TEST_EMBEDDED_OWN", "mine")

	var cfg embeddingConfig
	err := Load(&cfg)
	if err == nil || !strings.Contains(err.Error(), "TEST_EMBEDDED_REQUIRED is required") {
		t.Fatalf("load error = %v", err)
	}
}
