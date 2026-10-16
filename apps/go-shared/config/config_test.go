package config

import (
	"strings"
	"testing"
	"time"
)

type testConfig struct {
	Port         string        `key:"TEST_PORT" default:"8080"`
	DatabaseURL  string        `key:"TEST_DATABASE_URL"`
	CookieSecure bool          `key:"TEST_COOKIE_SECURE" default:"false"`
	Timeout      time.Duration `key:"TEST_TIMEOUT" default:"5m"`
	BatchSize    int           `key:"TEST_BATCH_SIZE" default:"10" min:"1" max:"100"`
	Ratio        float64       `key:"TEST_RATIO" default:"0.5" min:"0" max:"1"`
	Signing      string        `key:"TEST_SIGNING_SECRET" minLength:"32"`
	Ignored      string
}

func TestLoadValidatesIntegerBounds(t *testing.T) {
	t.Setenv("TEST_DATABASE_URL", "postgres://localhost/db")
	t.Setenv("TEST_SIGNING_SECRET", strings.Repeat("a", 32))
	t.Setenv("TEST_BATCH_SIZE", "101")

	var cfg testConfig
	err := Load(&cfg)
	if err == nil || !strings.Contains(err.Error(), "TEST_BATCH_SIZE must be at most 100") {
		t.Fatalf("load error = %v", err)
	}
}

func TestLoadReadsAFloatWithinBounds(t *testing.T) {
	t.Setenv("TEST_DATABASE_URL", "postgres://localhost/db")
	t.Setenv("TEST_SIGNING_SECRET", strings.Repeat("a", 32))

	t.Setenv("TEST_RATIO", "0.65")
	var cfg testConfig
	if err := Load(&cfg); err != nil || cfg.Ratio != 0.65 {
		t.Fatalf("ratio = %v, err = %v", cfg.Ratio, err)
	}

	t.Setenv("TEST_RATIO", "1.2")
	if err := Load(&cfg); err == nil || !strings.Contains(err.Error(), "TEST_RATIO must be at most 1") {
		t.Errorf("load error = %v", err)
	}
}

func TestLoadReadsTheEnvironmentAndFallsBackToDefaults(t *testing.T) {
	secret := strings.Repeat("a", 32)
	t.Setenv("TEST_DATABASE_URL", "postgres://localhost/db")
	t.Setenv("TEST_COOKIE_SECURE", "true")
	t.Setenv("TEST_SIGNING_SECRET", secret)

	var cfg testConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected the default port, got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://localhost/db" {
		t.Errorf("expected the environment's database url, got %q", cfg.DatabaseURL)
	}
	if !cfg.CookieSecure {
		t.Error("expected COOKIE_SECURE=true to parse as true")
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("expected the default timeout, got %s", cfg.Timeout)
	}
	if cfg.Signing != secret {
		t.Error("expected the signing secret to be loaded")
	}
	if cfg.Ignored != "" {
		t.Errorf("expected a field without an env tag to be left alone, got %q", cfg.Ignored)
	}
}

func TestLoadReportsEveryProblemAtOnce(t *testing.T) {
	t.Setenv("TEST_COOKIE_SECURE", "yes-please")
	t.Setenv("TEST_TIMEOUT", "eventually")
	t.Setenv("TEST_SIGNING_SECRET", "too-short")

	var cfg testConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("expected the three bad variables to be reported")
	}

	for _, want := range []string{
		"TEST_DATABASE_URL is required",
		"TEST_COOKIE_SECURE must be true or false",
		"TEST_TIMEOUT must be a duration such as 30s or 5m",
		"TEST_SIGNING_SECRET must be at least 32 characters",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected %q in %q", want, err)
		}
	}

	if strings.Contains(err.Error(), "too-short") {
		t.Errorf("a configured value leaked into the error: %q", err)
	}
	if strings.Contains(err.Error(), "TEST_PORT") {
		t.Errorf("expected a defaulted variable not to be reported: %q", err)
	}
}

func TestLoadRejectsANonPointer(t *testing.T) {
	if err := Load(testConfig{}); err == nil {
		t.Error("expected loading into a non-pointer to fail")
	}
}
